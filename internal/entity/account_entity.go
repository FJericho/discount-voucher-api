package entity

import "time"

const (
	ADMIN    = "admin"
	USER     = "user"
)

type Account struct {
	ID         string       `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name       string       `gorm:"not null"`
	Email      string       `gorm:"uniqueIndex;not null"`
	Password   string       `gorm:"not null"`
	Role       string       `gorm:"default:'user'"`
	CreatedAt  time.Time    `gorm:"default:current_timestamp"`
}
