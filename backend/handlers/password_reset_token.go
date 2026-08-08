package handlers

import "time"

type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey"`
	Email     string    `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt int64     `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func createPasswordResetToken(email string, token string, expiresAt int64) *PasswordResetToken {
	return &PasswordResetToken{
		Email:     email,
		Token:     token,
		ExpiresAt: expiresAt,
	}
}
