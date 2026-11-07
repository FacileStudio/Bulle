package computers

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/worker"
	"github.com/FacileStudio/tronc/errors"
)

// The egress modes a computer can run in. Offline denies the guest every
// outbound connection on the host side; SSH into it keeps working.
const (
	NetworkOpen    = "open"
	NetworkOffline = "offline"
)

// Networks reads and saves the egress mode each person chose for their
// computer. A person who never chose reads as NetworkOpen.
type Networks interface {
	ComputerNetwork(ctx context.Context, userID int64) (string, error)
	SaveComputerNetwork(ctx context.Context, userID int64, network string) error
}

// ValidNetwork reports whether network is one of the supported egress modes.
func ValidNetwork(network string) bool {
	return network == NetworkOpen || network == NetworkOffline
}

func (s *Service) network(ctx context.Context, userID int64) (string, error) {
	if s.Networks == nil {
		return NetworkOpen, nil
	}
	network, err := s.Networks.ComputerNetwork(ctx, userID)
	if err != nil {
		return "", errors.Internal("failed to load the computer network setting", err)
	}
	if network == "" {
		return NetworkOpen, nil
	}
	return network, nil
}

func (s *Service) withNetwork(ctx context.Context, computer *Computer, err error) (*Computer, error) {
	if err != nil {
		return nil, err
	}
	network, err := s.network(ctx, computer.UserID)
	if err != nil {
		return nil, err
	}
	computer.Network = network
	return computer, nil
}

func boxNetwork(box *worker.Box) string {
	if box.Net == "" {
		return NetworkOpen
	}
	return box.Net
}

// wireNetwork leaves the open default off the create request, so a worker
// that predates egress modes still accepts it, while an offline request to
// such a worker fails instead of silently booting open.
func wireNetwork(network string) string {
	if network == NetworkOpen {
		return ""
	}
	return network
}
