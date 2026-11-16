package computers

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
	"github.com/FacileStudio/tronc/errors"
)

// SetNetwork saves the person's egress mode and hands it to their computer.
// The worker reads the mode at boot, so a running computer restarts to take
// it and a stopped one takes it on its next start.
//
// allow replaces the person's extra hosts and is only accepted with the
// allowlist mode; leaving it out clears them. The hosts are checked with the
// worker's rules first, so a bad entry is a 400 here.
func (s *Service) SetNetwork(ctx context.Context, userID int64, network string, allow []string) (*Computer, error) {
	if !ValidNetwork(network) {
		return nil, errors.Invalid("network must be open, offline or allowlist")
	}
	if network != NetworkAllowlist && len(allow) > 0 {
		return nil, errors.Invalid("net_allow needs the allowlist network")
	}
	allow, err := ValidateNetAllow(allow)
	if err != nil {
		return nil, err
	}
	if s.Networks == nil {
		return nil, errors.Unavailable("the computer network setting is not available on this server")
	}
	defer s.locks.lock(userID)()
	if err := s.Networks.SaveComputerNetwork(ctx, userID, network, allow); err != nil {
		return nil, errors.Internal("failed to save the computer network setting", err)
	}
	computer, box, err := s.load(ctx, userID)
	if err != nil {
		return nil, err
	}
	if box == nil {
		return &Computer{UserID: userID, Status: StatusAbsent, Network: network, NetAllow: allow}, nil
	}
	box, err = s.applyNetwork(ctx, box, netChoice{mode: network, allow: allow})
	if err != nil {
		return nil, err
	}
	saved, err := s.persist(ctx, computer, box, userID, false)
	return s.withNetwork(ctx, saved, err)
}

func (s *Service) applyNetwork(ctx context.Context, box *worker.Box, choice netChoice) (*worker.Box, error) {
	if box.Status == worker.StatusCreating {
		settled, err := s.waitFor(ctx, box, func(b worker.Box) bool { return b.Status != worker.StatusCreating })
		if err != nil {
			return nil, err
		}
		box = settled
	}
	if box.Status == worker.StatusError || choice.matches(box) {
		return box, nil
	}
	if err := s.worker.SetNet(ctx, box.ID, choice.mode, choice.allow); err != nil {
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

func (s *Service) createBox(ctx context.Context, userID int64, choice netChoice) (*worker.Box, error) {
	request := worker.CreateRequest{
		Name:   boxName(userID),
		Owner:  ownerID(userID),
		CPUs:   s.config.CPUs,
		Memory: formatMemory(s.config.MemoryMB),
		Disk:   formatDisk(s.config.DiskGB),
		SSHKey: s.config.PublicKey,
		Net:    wireNetwork(choice.mode),
	}
	if choice.mode == NetworkAllowlist {
		request.NetAllow = choice.allow
	}
	fresh, err := s.worker.Create(ctx, request)
	if err != nil {
		return nil, err
	}
	return &fresh, nil
}

func (s *Service) boot(ctx context.Context, box *worker.Box, choice netChoice) (*worker.Box, error) {
	if err := s.alignNetwork(ctx, box, choice); err != nil {
		return nil, err
	}
	if err := s.worker.Start(ctx, box.ID); err != nil {
		return nil, err
	}
	return s.fetch(ctx, box.ID)
}

// alignNetwork gives a stopped box the person's egress mode before it boots,
// so a choice that never reached the worker still holds.
func (s *Service) alignNetwork(ctx context.Context, box *worker.Box, choice netChoice) error {
	if choice.matches(box) {
		return nil
	}
	return s.worker.SetNet(ctx, box.ID, choice.mode, choice.allow)
}

// keepDefaultNetwork saves the instance default as the choice of a person who
// never made one, once a computer was built with it, so a later change of the
// default leaves that computer's mode alone.
func (s *Service) keepDefaultNetwork(ctx context.Context, userID int64, choice netChoice) {
	if s.Networks == nil {
		return
	}
	mode, _, err := s.Networks.ComputerNetwork(ctx, userID)
	if err != nil || mode != "" {
		return
	}
	if err := s.Networks.SaveComputerNetwork(ctx, userID, choice.mode, choice.allow); err != nil {
		s.logger.Warn("failed to keep the default computer network", slog.Int64("user_id", userID), slog.Any("error", err))
	}
}
