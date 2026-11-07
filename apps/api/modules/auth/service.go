package auth

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Bulle/apps/api/internal/authcontext"
	"github.com/FacileStudio/Bulle/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

type Service struct {
	orm *gorm.DB
}

func NewService(orm *gorm.DB) *Service {
	return &Service{orm: orm}
}

func (s *Service) IdentityForUser(ctx context.Context, userID int64) (authcontext.Identity, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).First(&user, userID).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return authcontext.Identity{}, errors.Unauthorized("invalid auth token")
	}
	if err != nil {
		return authcontext.Identity{}, errors.Internal("failed to load the account", err)
	}
	return authcontext.Identity{UserID: user.ID, Email: user.Email}, nil
}

func (s *Service) UserByID(ctx context.Context, id int64) (*schemas.User, error) {
	var user schemas.User
	if err := s.orm.WithContext(ctx).First(&user, id).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal("failed to load user", err)
	}
	return &user, nil
}

// IdentityForEmail finds the account an email address belongs to. It answers
// NotFound when no Bulle user holds it.
func (s *Service) IdentityForEmail(ctx context.Context, email string) (authcontext.Identity, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).Where("email = ?", strings.ToLower(strings.TrimSpace(email))).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return authcontext.Identity{}, errors.NotFound("user not found")
	}
	if err != nil {
		return authcontext.Identity{}, errors.Internal("failed to load the account", err)
	}
	return authcontext.Identity{UserID: user.ID, Email: user.Email}, nil
}
