package computers

import (
	"context"
	"slices"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
	"github.com/FacileStudio/tronc/errors"
)

// The egress modes a computer can run in. Offline denies the guest every
// outbound connection on the host side; SSH into it keeps working. Allowlist
// reaches the worker's built-in package registries and GitHub, plus the
// person's extra hosts, through a host-side proxy.
const (
	NetworkOpen      = "open"
	NetworkOffline   = "offline"
	NetworkAllowlist = "allowlist"
)

// Networks reads and saves the egress mode each person chose for their
// computer and, for allowlist, their extra hosts. An empty mode means the
// person never chose.
type Networks interface {
	ComputerNetwork(ctx context.Context, userID int64) (mode string, allow []string, err error)
	SaveComputerNetwork(ctx context.Context, userID int64, mode string, allow []string) error
}

// ValidNetwork reports whether network is one of the supported egress modes.
func ValidNetwork(network string) bool {
	return network == NetworkOpen || network == NetworkOffline || network == NetworkAllowlist
}

// netChoice is the egress mode a computer runs in and the extra hosts that go
// with it.
type netChoice struct {
	mode  string
	allow []string
}

// network is the person's effective choice. A person who never chose gets the
// instance default for a computer they do not have yet, and open for one that
// already exists, so a computer built before the allowlist keeps its mode.
func (s *Service) network(ctx context.Context, userID int64, existing bool) (netChoice, error) {
	if s.Networks == nil {
		return netChoice{mode: NetworkOpen}, nil
	}
	mode, allow, err := s.Networks.ComputerNetwork(ctx, userID)
	if err != nil {
		return netChoice{}, errors.Internal("failed to load the computer network setting", err)
	}
	if mode == "" {
		mode = NetworkOpen
		if !existing && s.config.DefaultNetwork == NetworkAllowlist {
			mode = NetworkAllowlist
		}
	}
	if mode != NetworkAllowlist {
		allow = nil
	}
	return netChoice{mode: mode, allow: allow}, nil
}

// NetworkOf is the egress mode and extra hosts of the person's computer, read
// from the database alone so a run can ask at its start. A failed read answers
// open, the mode that asks the most of the gate.
func (s *Service) NetworkOf(ctx context.Context, userID int64) (string, []string) {
	computer, err := s.store.Find(ctx, userID)
	if err != nil {
		return NetworkOpen, nil
	}
	choice, err := s.network(ctx, userID, computer != nil)
	if err != nil {
		return NetworkOpen, nil
	}
	return choice.mode, choice.allow
}

func (s *Service) withNetwork(ctx context.Context, computer *Computer, err error) (*Computer, error) {
	if err != nil {
		return nil, err
	}
	choice, err := s.network(ctx, computer.UserID, computer.BoxID != "")
	if err != nil {
		return nil, err
	}
	computer.Network = choice.mode
	computer.NetAllow = choice.allow
	return computer, nil
}

func boxNetwork(box *worker.Box) string {
	if box.Net == "" {
		return NetworkOpen
	}
	return box.Net
}

func (c netChoice) matches(box *worker.Box) bool {
	if boxNetwork(box) != c.mode {
		return false
	}
	have, want := slices.Clone(box.NetAllow), slices.Clone(c.allow)
	slices.Sort(have)
	slices.Sort(want)
	return slices.Equal(have, want)
}

// wireNetwork leaves the open default off the create request, so a worker
// that predates egress modes still accepts it, while an offline or allowlist
// request to such a worker fails instead of silently booting open.
func wireNetwork(network string) string {
	if network == NetworkOpen {
		return ""
	}
	return network
}
