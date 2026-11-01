package computers

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
	"github.com/FacileStudio/tronc/errors"
)

var tagPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,63}$`)

type Worker interface {
	Create(ctx context.Context, request worker.CreateRequest) (worker.Box, error)
	List(ctx context.Context, owner string) ([]worker.Box, error)
	Get(ctx context.Context, id string) (worker.Box, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Snapshot(ctx context.Context, id, tag string) error
	Rollback(ctx context.Context, id, tag string) error
}

type Service struct {
	store  Store
	worker Worker
	config Config
	logger *slog.Logger
	locks  userLocks
	Probe  func(ctx context.Context, host string, port int) error
}

func NewService(store Store, boxes Worker, config Config, logger *slog.Logger) *Service {
	return &Service{store: store, worker: boxes, config: config, logger: logger, Probe: probeSSH}
}

func (s *Service) Ensure(ctx context.Context, userID int64) (*Computer, error) {
	return s.converge(ctx, userID, true)
}

func (s *Service) Launch(ctx context.Context, userID int64) (*Computer, error) {
	return s.converge(ctx, userID, false)
}

func (s *Service) Status(ctx context.Context, userID int64) (*Computer, error) {
	defer s.locks.lock(userID)()
	computer, box, err := s.load(ctx, userID)
	if err != nil {
		return nil, err
	}
	if box == nil {
		return &Computer{UserID: userID, Status: StatusAbsent}, nil
	}
	return s.persist(ctx, computer, box, userID, false)
}

func (s *Service) Stop(ctx context.Context, userID int64) (*Computer, error) {
	defer s.locks.lock(userID)()
	return s.stopLocked(ctx, userID, false)
}

func (s *Service) Snapshot(ctx context.Context, userID int64, tag string) (*Computer, error) {
	return s.offline(ctx, userID, tag, s.worker.Snapshot)
}

func (s *Service) Rollback(ctx context.Context, userID int64, tag string) (*Computer, error) {
	return s.offline(ctx, userID, tag, s.worker.Rollback)
}

func (s *Service) Touch(ctx context.Context, userID int64) error {
	return s.store.Touch(ctx, userID, time.Now())
}

func (s *Service) offline(ctx context.Context, userID int64, tag string, act func(context.Context, string, string) error) (*Computer, error) {
	if !tagPattern.MatchString(tag) {
		return nil, errors.Invalid("tag must be 1 to 64 letters, digits, dots, dashes or underscores")
	}
	defer s.locks.lock(userID)()
	computer, err := s.stopLocked(ctx, userID, true)
	if err != nil {
		return nil, err
	}
	if computer.Status == StatusAbsent {
		return nil, errors.NotFound("computer not found")
	}
	if err := act(ctx, computer.BoxID, tag); err != nil {
		return nil, err
	}
	return computer, nil
}

func (s *Service) stopLocked(ctx context.Context, userID int64, wait bool) (*Computer, error) {
	computer, box, err := s.load(ctx, userID)
	if err != nil {
		return nil, err
	}
	if box == nil {
		return &Computer{UserID: userID, Status: StatusAbsent}, nil
	}
	if box.Status == worker.StatusCreating && wait {
		settled, err := s.waitFor(ctx, box, func(b worker.Box) bool { return b.Status != worker.StatusCreating })
		if err != nil {
			return nil, err
		}
		box = settled
	}
	if box.Status == worker.StatusRunning {
		if err := s.worker.Stop(ctx, box.ID); err != nil {
			return nil, err
		}
		stopped, err := s.worker.Get(ctx, box.ID)
		if err != nil {
			return nil, err
		}
		box = &stopped
	}
	if wait {
		waited, err := s.waitFor(ctx, box, func(b worker.Box) bool { return b.Status == worker.StatusStopped })
		if err != nil {
			return nil, err
		}
		box = waited
	}
	return s.persist(ctx, computer, box, userID, false)
}

func (s *Service) waitFor(ctx context.Context, box *worker.Box, done func(worker.Box) bool) (*worker.Box, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.BootTimeout)
	defer cancel()
	for !done(*box) {
		if box.Status == worker.StatusError {
			return nil, errors.Failed("the computer is in an error state: " + box.Error)
		}
		if err := sleep(ctx, s.config.PollInterval); err != nil {
			return nil, errors.Unavailable("the computer did not reach the expected state in time")
		}
		next, err := s.worker.Get(ctx, box.ID)
		if err != nil {
			return nil, err
		}
		box = &next
	}
	return box, nil
}

func (s *Service) converge(ctx context.Context, userID int64, wait bool) (*Computer, error) {
	defer s.locks.lock(userID)()
	computer, box, err := s.load(ctx, userID)
	if err != nil {
		return nil, err
	}
	if box == nil {
		created, err := s.worker.Create(ctx, worker.CreateRequest{
			Name:   boxName(userID),
			Owner:  ownerID(userID),
			CPUs:   s.config.CPUs,
			Memory: formatMemory(s.config.MemoryMB),
			Disk:   formatDisk(s.config.DiskGB),
		})
		if err != nil {
			return nil, err
		}
		box = &created
	}
	if box.Status == worker.StatusStopped {
		if err := s.worker.Start(ctx, box.ID); err != nil {
			return nil, err
		}
		started, err := s.worker.Get(ctx, box.ID)
		if err != nil {
			return nil, err
		}
		box = &started
	}
	if wait {
		ready, err := s.waitReady(ctx, box)
		if err != nil {
			return nil, err
		}
		box = ready
	}
	return s.persist(ctx, computer, box, userID, true)
}

func (s *Service) waitReady(ctx context.Context, box *worker.Box) (*worker.Box, error) {
	ctx, cancel := context.WithTimeout(ctx, s.config.BootTimeout)
	defer cancel()
	for {
		switch box.Status {
		case worker.StatusError:
			return nil, errors.Failed("the computer failed to start: " + box.Error)
		case worker.StatusRunning:
			if s.Probe(ctx, box.SSH.Host, box.SSH.Port) == nil {
				return box, nil
			}
		}
		if err := sleep(ctx, s.config.PollInterval); err != nil {
			return nil, errors.Unavailable("the computer did not become ready in time")
		}
		next, err := s.worker.Get(ctx, box.ID)
		if err != nil {
			return nil, err
		}
		box = &next
	}
}

func (s *Service) load(ctx context.Context, userID int64) (*Computer, *worker.Box, error) {
	computer, err := s.store.Find(ctx, userID)
	if err != nil {
		return nil, nil, errors.Internal("failed to load the computer", err)
	}
	if computer != nil && computer.BoxID != "" {
		box, err := s.worker.Get(ctx, computer.BoxID)
		if err == nil {
			return computer, &box, nil
		}
		if !worker.IsCode(err, worker.CodeNotFound) {
			return nil, nil, err
		}
		s.logger.Warn("computer row points at a missing box, dropping it", slog.Int64("user_id", userID), slog.String("box_id", computer.BoxID))
		if err := s.store.Delete(ctx, userID); err != nil {
			return nil, nil, errors.Internal("failed to reconcile the computer", err)
		}
		computer = nil
	}
	boxes, err := s.worker.List(ctx, ownerID(userID))
	if err != nil {
		return nil, nil, err
	}
	for i := range boxes {
		if boxes[i].Owner == ownerID(userID) {
			return computer, &boxes[i], nil
		}
	}
	return computer, nil, nil
}

func (s *Service) persist(ctx context.Context, computer *Computer, box *worker.Box, userID int64, active bool) (*Computer, error) {
	if computer == nil {
		computer = &Computer{UserID: userID, LastActiveAt: time.Now()}
	}
	computer.BoxID = box.ID
	computer.Status = box.Status
	computer.SSHHost = box.SSH.Host
	computer.SSHPort = box.SSH.Port
	computer.SSHUser = box.SSH.User
	computer.CPUs = box.CPUs
	computer.MemoryMB = parseMemoryMB(box.Memory, s.config.MemoryMB)
	computer.DiskGB = parseDiskGB(box.Disk, s.config.DiskGB)
	computer.Error = box.Error
	if active {
		computer.LastActiveAt = time.Now()
	}
	if err := s.store.Save(ctx, computer); err != nil {
		return nil, errors.Internal("failed to save the computer", err)
	}
	return computer, nil
}

func formatMemory(mb int) string {
	if mb <= 0 {
		return "4096M"
	}
	return strconv.Itoa(mb) + "M"
}

func formatDisk(gb int) string {
	if gb <= 0 {
		return "20G"
	}
	return strconv.Itoa(gb) + "G"
}

func parseMemoryMB(raw string, fallback int) int {
	raw = strings.TrimSpace(strings.ToUpper(raw))
	if strings.HasSuffix(raw, "G") {
		if val, err := strconv.Atoi(strings.TrimSuffix(raw, "G")); err == nil {
			return val * 1024
		}
	}
	if strings.HasSuffix(raw, "M") {
		if val, err := strconv.Atoi(strings.TrimSuffix(raw, "M")); err == nil {
			return val
		}
	}
	return fallback
}

func parseDiskGB(raw string, fallback int) int {
	raw = strings.TrimSpace(strings.ToUpper(raw))
	raw = strings.TrimSuffix(raw, "G")
	if val, err := strconv.Atoi(raw); err == nil {
		return val
	}
	return fallback
}

func boxName(userID int64) string { return "bulle-" + strconv.FormatInt(userID, 36) }

func ownerID(userID int64) string { return strconv.FormatInt(userID, 10) }

type userLocks struct {
	mu    sync.Mutex
	locks map[int64]*sync.Mutex
}

func (u *userLocks) lock(userID int64) func() {
	u.mu.Lock()
	if u.locks == nil {
		u.locks = map[int64]*sync.Mutex{}
	}
	mu, ok := u.locks[userID]
	if !ok {
		mu = &sync.Mutex{}
		u.locks[userID] = mu
	}
	u.mu.Unlock()
	mu.Lock()
	return mu.Unlock
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func probeSSH(ctx context.Context, host string, port int) error {
	if host == "" || port <= 0 {
		return fmt.Errorf("no ssh endpoint yet")
	}
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	var dialer net.Dialer
	conn, err := dialer.DialContext(dialCtx, "tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	banner, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(banner, "SSH-") {
		return fmt.Errorf("not an ssh server")
	}
	return nil
}
