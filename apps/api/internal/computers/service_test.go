package computers

import (
	"context"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/tronc/errors"
)

type memoryStore struct {
	mu   sync.Mutex
	rows map[int64]Computer
}

func newMemoryStore() *memoryStore { return &memoryStore{rows: map[int64]Computer{}} }

func (m *memoryStore) Find(_ context.Context, userID int64) (*Computer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.rows[userID]
	if !ok {
		return nil, nil
	}
	return &row, nil
}

func (m *memoryStore) Save(_ context.Context, c *Computer) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rows[c.UserID] = *c
	return nil
}

func (m *memoryStore) Delete(_ context.Context, userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.rows, userID)
	return nil
}

func (m *memoryStore) Touch(_ context.Context, userID int64, at time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	row := m.rows[userID]
	row.LastActiveAt = at
	m.rows[userID] = row
	return nil
}

func (m *memoryStore) Idle(_ context.Context, before time.Time) ([]Computer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []Computer
	for _, row := range m.rows {
		if row.Status == StatusRunning && row.LastActiveAt.Before(before) {
			out = append(out, row)
		}
	}
	return out, nil
}

type fakeWorker struct {
	mu       sync.Mutex
	boxes    map[string]*worker.Box
	next     int
	creates  int
	calls    []string
	settleAt int
	polls    map[string]int
}

func newFakeWorker() *fakeWorker {
	return &fakeWorker{boxes: map[string]*worker.Box{}, polls: map[string]int{}, settleAt: 2}
}

func (f *fakeWorker) Create(_ context.Context, r worker.CreateRequest) (worker.Box, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	f.creates++
	box := &worker.Box{ID: "box" + string(rune('0'+f.next)), Name: r.Name, Owner: r.Owner, Status: worker.StatusCreating,
		CPUs: r.CPUs, Memory: r.Memory, Disk: r.Disk, SSH: worker.SSH{Host: "127.0.0.1", Port: 2222, User: "boite"}}
	f.boxes[box.ID] = box
	return *box, nil
}

func (f *fakeWorker) List(_ context.Context, owner string) ([]worker.Box, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []worker.Box
	for _, b := range f.boxes {
		if b.Owner == owner {
			out = append(out, *b)
		}
	}
	return out, nil
}

func (f *fakeWorker) Get(_ context.Context, id string) (worker.Box, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	box, ok := f.boxes[id]
	if !ok {
		return worker.Box{}, errors.New("not_found", "gone", &worker.Error{Code: worker.CodeNotFound})
	}
	if box.Status == worker.StatusCreating {
		f.polls[id]++
		if f.polls[id] >= f.settleAt {
			box.Status = worker.StatusRunning
		}
	}
	return *box, nil
}

func (f *fakeWorker) Start(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, "start")
	f.polls[id] = 0
	f.boxes[id].Status = worker.StatusCreating
	return nil
}

func (f *fakeWorker) Stop(_ context.Context, id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, "stop")
	f.boxes[id].Status = worker.StatusStopped
	return nil
}

func (f *fakeWorker) Snapshot(_ context.Context, id, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.boxes[id].Status != worker.StatusStopped {
		return errors.Failed("not stopped")
	}
	f.calls = append(f.calls, "snapshot:"+tag)
	return nil
}

func (f *fakeWorker) Rollback(_ context.Context, id, tag string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.boxes[id].Status != worker.StatusStopped {
		return errors.Failed("not stopped")
	}
	f.calls = append(f.calls, "rollback:"+tag)
	return nil
}

func newTestService(t *testing.T) (*Service, *fakeWorker, *memoryStore) {
	t.Helper()
	fw := newFakeWorker()
	store := newMemoryStore()
	service := NewService(store, fw, Config{
		CPUs: 2, MemoryMB: 4096, DiskGB: 20,
		IdleAfter: 30 * time.Minute, BootTimeout: 5 * time.Second, PollInterval: time.Millisecond,
		KeyPath: "/keys/box",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	service.Probe = func(context.Context, string, int) error { return nil }
	return service, fw, store
}

func TestEnsureCreatesTheBoxOnFirstUseAndWaitsForRunning(t *testing.T) {
	service, fw, store := newTestService(t)
	computer, err := service.Ensure(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if computer.Status != StatusRunning || computer.SSHPort != 2222 || computer.CPUs != 2 {
		t.Fatalf("computer = %+v", computer)
	}
	if fw.creates != 1 {
		t.Fatalf("creates = %d", fw.creates)
	}
	box := fw.boxes["box1"]
	if box.Name != "bulle-16" || box.Owner != "42" {
		t.Fatalf("box = %+v", box)
	}
	if row, _ := store.Find(context.Background(), 42); row == nil || row.BoxID != "box1" {
		t.Fatalf("row = %+v", row)
	}
	if _, err := service.Ensure(context.Background(), 42); err != nil || fw.creates != 1 {
		t.Fatalf("second ensure created again: creates=%d err=%v", fw.creates, err)
	}
}

func TestEnsureStartsAStoppedBox(t *testing.T) {
	service, fw, _ := newTestService(t)
	if _, err := service.Ensure(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Stop(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	computer, err := service.Ensure(context.Background(), 1)
	if err != nil || computer.Status != StatusRunning {
		t.Fatalf("computer = %+v err = %v", computer, err)
	}
	if got := fw.calls; len(got) != 2 || got[0] != "stop" || got[1] != "start" {
		t.Fatalf("calls = %v", got)
	}
}

func TestEnsureWaitsForSSH(t *testing.T) {
	service, _, _ := newTestService(t)
	attempts := 0
	service.Probe = func(context.Context, string, int) error {
		attempts++
		if attempts < 3 {
			return io.EOF
		}
		return nil
	}
	if _, err := service.Ensure(context.Background(), 1); err != nil || attempts != 3 {
		t.Fatalf("attempts = %d err = %v", attempts, err)
	}
}

func TestEnsureTimesOutWhenSSHNeverAnswers(t *testing.T) {
	service, _, _ := newTestService(t)
	service.config.BootTimeout = 30 * time.Millisecond
	service.Probe = func(context.Context, string, int) error { return io.EOF }
	_, err := service.Ensure(context.Background(), 1)
	if errors.Status(err) != 503 {
		t.Fatalf("err = %v", err)
	}
}

func TestEnsureReportsAnErroredBox(t *testing.T) {
	service, fw, _ := newTestService(t)
	fw.settleAt = 1 << 30
	go func() {
		for {
			fw.mu.Lock()
			if b, ok := fw.boxes["box1"]; ok {
				b.Status = worker.StatusError
				b.Error = "disk full"
				fw.mu.Unlock()
				return
			}
			fw.mu.Unlock()
			time.Sleep(time.Millisecond)
		}
	}()
	_, err := service.Ensure(context.Background(), 1)
	if errors.Status(err) != 412 {
		t.Fatalf("err = %v", err)
	}
}

func TestReconcileDropsARowWhoseBoxVanished(t *testing.T) {
	service, fw, store := newTestService(t)
	_ = store.Save(context.Background(), &Computer{UserID: 5, BoxID: "ghost", Status: StatusRunning})
	computer, err := service.Status(context.Background(), 5)
	if err != nil || computer.Status != StatusAbsent {
		t.Fatalf("computer = %+v err = %v", computer, err)
	}
	if row, _ := store.Find(context.Background(), 5); row != nil {
		t.Fatalf("stale row survived: %+v", row)
	}
	if _, err := service.Ensure(context.Background(), 5); err != nil || fw.creates != 1 {
		t.Fatalf("recreate failed: creates=%d err=%v", fw.creates, err)
	}
}

func TestReconcileAdoptsAnOrphanBox(t *testing.T) {
	service, fw, store := newTestService(t)
	fw.boxes["orphan"] = &worker.Box{ID: "orphan", Owner: "8", Status: worker.StatusStopped, CPUs: 1}
	computer, err := service.Status(context.Background(), 8)
	if err != nil || computer.Status != StatusStopped || computer.BoxID != "orphan" {
		t.Fatalf("computer = %+v err = %v", computer, err)
	}
	if row, _ := store.Find(context.Background(), 8); row == nil || row.BoxID != "orphan" {
		t.Fatalf("row = %+v", row)
	}
	if _, err := service.Ensure(context.Background(), 8); err != nil || fw.creates != 0 {
		t.Fatalf("creates = %d err = %v", fw.creates, err)
	}
}

func TestStatusIsAbsentWithoutAComputer(t *testing.T) {
	service, _, _ := newTestService(t)
	computer, err := service.Status(context.Background(), 3)
	if err != nil || computer.Status != StatusAbsent {
		t.Fatalf("computer = %+v err = %v", computer, err)
	}
}

func TestUsersOnlyReachTheirOwnBox(t *testing.T) {
	service, fw, _ := newTestService(t)
	if _, err := service.Ensure(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	other, err := service.Status(context.Background(), 2)
	if err != nil || other.Status != StatusAbsent {
		t.Fatalf("user 2 saw %+v err = %v", other, err)
	}
	if _, err := service.Ensure(context.Background(), 2); err != nil || fw.creates != 2 {
		t.Fatalf("creates = %d err = %v", fw.creates, err)
	}
}

func TestSnapshotAndRollbackStopFirst(t *testing.T) {
	service, fw, _ := newTestService(t)
	if _, err := service.Ensure(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	computer, err := service.Snapshot(context.Background(), 1, "v1")
	if err != nil || computer.Status != StatusStopped {
		t.Fatalf("computer = %+v err = %v", computer, err)
	}
	if _, err := service.Rollback(context.Background(), 1, "v1"); err != nil {
		t.Fatal(err)
	}
	want := []string{"stop", "snapshot:v1", "rollback:v1"}
	for i, call := range want {
		if i >= len(fw.calls) || fw.calls[i] != call {
			t.Fatalf("calls = %v, want %v", fw.calls, want)
		}
	}
}

func TestSnapshotRejectsBadTagsAndMissingComputers(t *testing.T) {
	service, _, _ := newTestService(t)
	if _, err := service.Snapshot(context.Background(), 1, "../x"); errors.Status(err) != 400 {
		t.Fatalf("bad tag err = %v", err)
	}
	if _, err := service.Snapshot(context.Background(), 1, "ok"); errors.Status(err) != 404 {
		t.Fatalf("missing computer err = %v", err)
	}
}

func TestSweepStopsOnlyIdleBoxes(t *testing.T) {
	service, fw, store := newTestService(t)
	for _, id := range []int64{1, 2} {
		if _, err := service.Ensure(context.Background(), id); err != nil {
			t.Fatal(err)
		}
	}
	row, _ := store.Find(context.Background(), 1)
	row.LastActiveAt = time.Now().Add(-time.Hour)
	_ = store.Save(context.Background(), row)

	if stopped := service.Sweep(context.Background()); stopped != 1 {
		t.Fatalf("stopped = %d", stopped)
	}
	if fw.boxes["box1"].Status != worker.StatusStopped || fw.boxes["box2"].Status != worker.StatusRunning {
		t.Fatalf("box1=%s box2=%s", fw.boxes["box1"].Status, fw.boxes["box2"].Status)
	}
	if row, _ := store.Find(context.Background(), 1); row.Status != StatusStopped {
		t.Fatalf("row status = %s", row.Status)
	}
}

func TestTouchKeepsABoxFromBeingSwept(t *testing.T) {
	service, fw, store := newTestService(t)
	if _, err := service.Ensure(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	row, _ := store.Find(context.Background(), 1)
	row.LastActiveAt = time.Now().Add(-time.Hour)
	_ = store.Save(context.Background(), row)

	NewProvider(service).Touch("1")
	if stopped := service.Sweep(context.Background()); stopped != 0 {
		t.Fatalf("stopped = %d", stopped)
	}
	if fw.boxes["box1"].Status != worker.StatusRunning {
		t.Fatal("touched box was stopped")
	}
}

func TestProviderBuildsASandboxTarget(t *testing.T) {
	service, _, _ := newTestService(t)
	var provider Provider = NewProvider(service)
	target, err := provider.Ensure(context.Background(), "42")
	if err != nil {
		t.Fatal(err)
	}
	want := sandbox.Target{
		Name: "bulle-16", Backend: "boite", Host: "127.0.0.1", Port: 2222, User: "boite",
		KeyPath: "/keys/box", Workdir: sandbox.GuestWorkspace, Status: StatusRunning,
	}
	if *target != want {
		t.Fatalf("target = %+v", *target)
	}
	if _, err := provider.Ensure(context.Background(), "abc"); errors.Status(err) != 400 {
		t.Fatalf("bad id err = %v", err)
	}
}

func TestConcurrentEnsureCreatesOneBox(t *testing.T) {
	service, fw, _ := newTestService(t)
	var wg sync.WaitGroup
	for range 5 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = service.Ensure(context.Background(), 9)
		}()
	}
	wg.Wait()
	if fw.creates != 1 {
		t.Fatalf("creates = %d", fw.creates)
	}
}

func TestResolveKeyStagesInlineKeysAndPassesPaths(t *testing.T) {
	if got, _ := resolveKey("/etc/key"); got != "/etc/key" {
		t.Fatalf("path = %q", got)
	}
	path, err := resolveKey("-----BEGIN OPENSSH PRIVATE KEY-----\nabc\n-----END OPENSSH PRIVATE KEY-----")
	if err != nil || path == "" {
		t.Fatalf("path = %q err = %v", path, err)
	}
}
