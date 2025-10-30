package db

import (
	"github.com/google/uuid"
	"time"
)

type Registration struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Username       string    `gorm:"type:varchar(100);not null"`
	Login          string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	HashedPassword string    `gorm:"type:varchar(255);not null"`
	CreatedAt      time.Time `gorm:"type:timestamp;not null"`
	UpdatedAt      time.Time `gorm:"type:timestamp;not null"`
}

type RefreshToken struct {
	TokenHash     []byte     `gorm:"primaryKey;column:token_hash;type:bytea"`
	UserID        uuid.UUID  `gorm:"not null;column:user_id;type:uuid;index"`
	IssuedAt      time.Time  `gorm:"not null;column:issued_at;default:now()"`
	ExpiresAt     time.Time  `gorm:"not null;column:expires_at"`
	RevokedAt     *time.Time `gorm:"column:revoked_at"`
	RevokedReason *string    `gorm:"column:revoked_reason"`
	CreatedAt     time.Time  `gorm:"primaryKey;not null;column:created_at;default:now()"`
}
