package computers

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
	"github.com/FacileStudio/tronc/errors"
)

// SetNetwork saves the person's egress mode and hands it to their computer.
// The worker reads the mode at boot, so a running computer restarts to take
// it and a stopped one takes it on its next start.
func (s *Service) SetNetwork(ctx context.Context, userID int64, network string) (*Computer, error) {
	if !ValidNetwork(network) {
		return nil, errors.Invalid("network must be open or offline")
	}
	if s.Networks == nil {
		return nil, errors.Unavailable("the computer network setting is not available on this server")
	}
	defer s.locks.lock(userID)()
	if err := s.Networks.SaveComputerNetwork(ctx, userID, network); err != nil {
		return nil, errors.Internal("failed to save the computer network setting", err)
	}
	computer, box, err := s.load(ctx, userID)
	if err != nil {
		return nil, err
	}
	if box == nil {
		return &Computer{UserID: userID, Status: StatusAbsent, Network: network}, nil
	}
	box, err = s.applyNetwork(ctx, box, network)
	if err != nil {
		return nil, err
	}
	saved, err := s.persist(ctx, computer, box, userID, false)
	return s.withNetwork(ctx, saved, err)
}

func (s *Service) applyNetwork(ctx context.Context, box *worker.Box, network string) (*worker.Box, error) {
	if box.Status == worker.StatusCreating {
		settled, err := s.waitFor(ctx, box, func(b worker.Box) bool { return b.Status != worker.StatusCreating })
		if err != nil {
			return nil, err
		}
		box = settled
	}
	if box.Status == worker.StatusError || boxNetwork(box) == network {
		return box, nil
	}
	if err := s.worker.SetNet(ctx, box.ID, network); err != nil {
		return nil, err
	}
	if box.Status == worker.StatusRunning {
		return s.restart(ctx, box)
	}
	return s.fetch(ctx, box.ID)
}

func (s *Service) restart(ctx context.Context, box *worker.Box) (*worker.Box, error) {
	if err := s.worker.Stop(ctx, box.ID); err != nil {
		return nil, err
	}
	stopped, err := s.fetch(ctx, box.ID)
	if err != nil {
		return nil, err
	}
	if _, err := s.waitFor(ctx, stopped, func(b worker.Box) bool { return b.Status == worker.StatusStopped }); err != nil {
		return nil, err
	}
	if err := s.worker.Start(ctx, box.ID); err != nil {
		return nil, err
	}
	return s.fetch(ctx, box.ID)
}

func (s *Service) fetch(ctx context.Context, id string) (*worker.Box, error) {
	box, err := s.worker.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return &box, nil
}

func (s *Service) createBox(ctx context.Context, userID int64, network string) (*worker.Box, error) {
	fresh, err := s.worker.Create(ctx, worker.CreateRequest{
		Name:   boxName(userID),
		Owner:  ownerID(userID),
		CPUs:   s.config.CPUs,
		Memory: formatMemory(s.config.MemoryMB),
		Disk:   formatDisk(s.config.DiskGB),
		SSHKey: s.config.PublicKey,
		Net:    wireNetwork(network),
	})
	if err != nil {
		return nil, err
	}
	return &fresh, nil
}

func (s *Service) boot(ctx context.Context, box *worker.Box, network string) (*worker.Box, error) {
	if err := s.alignNetwork(ctx, box, network); err != nil {
		return nil, err
	}
	if err := s.worker.Start(ctx, box.ID); err != nil {
		return nil, err
	}
	return s.fetch(ctx, box.ID)
}

// alignNetwork gives a stopped box the person's egress mode before it boots,
// so a choice that never reached the worker still holds.
func (s *Service) alignNetwork(ctx context.Context, box *worker.Box, network string) error {
	if boxNetwork(box) == network {
		return nil
	}
	return s.worker.SetNet(ctx, box.ID, network)
}
