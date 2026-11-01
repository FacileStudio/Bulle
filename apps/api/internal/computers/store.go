package computers

import (
	"context"
	stderrors "errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store interface {
	Find(ctx context.Context, userID int64) (*Computer, error)
	Save(ctx context.Context, computer *Computer) error
	Delete(ctx context.Context, userID int64) error
	Touch(ctx context.Context, userID int64, at time.Time) error
	Idle(ctx context.Context, before time.Time) ([]Computer, error)
}

type GormStore struct {
	orm *gorm.DB
}

func NewStore(orm *gorm.DB) *GormStore {
	return &GormStore{orm: orm}
}

func (s *GormStore) Find(ctx context.Context, userID int64) (*Computer, error) {
	var computer Computer
	err := s.orm.WithContext(ctx).Where("user_id = ?", userID).First(&computer).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &computer, nil
}

func (s *GormStore) Save(ctx context.Context, computer *Computer) error {
	return s.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"box_id", "status", "ssh_host", "ssh_port", "cpus", "memory_mb", "disk_gb", "last_active_at",
		}),
	}).Create(computer).Error
}

func (s *GormStore) Delete(ctx context.Context, userID int64) error {
	return s.orm.WithContext(ctx).Where("user_id = ?", userID).Delete(&Computer{}).Error
}

func (s *GormStore) Touch(ctx context.Context, userID int64, at time.Time) error {
	return s.orm.WithContext(ctx).Model(&Computer{}).Where("user_id = ?", userID).
		Update("last_active_at", at).Error
}

func (s *GormStore) Idle(ctx context.Context, before time.Time) ([]Computer, error) {
	var idle []Computer
	err := s.orm.WithContext(ctx).
		Where("status = ? AND last_active_at < ?", StatusRunning, before).
		Find(&idle).Error
	return idle, err
}
