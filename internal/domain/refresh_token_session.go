package domain

import "time"

type RefreshTokenSession struct {
	ID                int64
	UserID            int64
	TokenID           string
	TokenHash         string
	IssuedAt          time.Time
	ExpiresAt         time.Time
	RevokedAt         *time.Time
	ReplacedByTokenID *string
	LastUsedAt        *time.Time
	CreatedAt         time.Time
}
