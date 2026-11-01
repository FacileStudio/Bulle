package computers

import (
	"context"
	"log/slog"
	"strconv"
	"time"

	"github.com/FacileStudio/Bulle/core/sandbox"
	"github.com/FacileStudio/tronc/errors"
)

type SandboxProvider struct {
	service *Service
}

func NewProvider(service *Service) *SandboxProvider {
	return &SandboxProvider{service: service}
}

var _ Provider = (*SandboxProvider)(nil)

func (p *SandboxProvider) Ensure(ctx context.Context, userID string) (*sandbox.Target, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	computer, err := p.service.Ensure(ctx, id)
	if err != nil {
		return nil, err
	}
	return p.service.Target(computer), nil
}

func (p *SandboxProvider) Touch(userID string) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := p.service.Touch(ctx, id); err != nil {
		p.service.logger.Warn("failed to record computer activity", slog.String("user_id", userID), slog.Any("error", err))
	}
}

func (s *Service) Target(computer *Computer) *sandbox.Target {
	return &sandbox.Target{
		Name:    boxName(computer.UserID),
		Backend: "boite",
		Host:    computer.SSHHost,
		Port:    computer.SSHPort,
		User:    computer.SSHUser,
		KeyPath: s.config.KeyPath,
		Workdir: sandbox.GuestWorkspace,
		Status:  computer.Status,
	}
}
