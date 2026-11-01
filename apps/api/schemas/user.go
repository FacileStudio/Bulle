package schemas

import "time"

type User struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Email     string    `gorm:"column:email;uniqueIndex"`
	Name      string    `gorm:"column:name"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (User) TableName() string { return "users" }
