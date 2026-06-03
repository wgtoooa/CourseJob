package auth

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"context"
	"time"
)

type RefreshTokenRepository struct {
	db contract.DBTX
}

func NewRefreshTokenRepository(db contract.DBTX) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, session *domain.RefreshTokenSession) error {
	const query = `
INSERT INTO auth_refresh_token (
	user_id,
	token_id,
	token_hash,
	issued_at,
	expires_at,
	revoked_at,
	replaced_by_token_id,
	last_used_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, created_at`

	return r.db.QueryRow(
		ctx,
		query,
		session.UserID,
		session.TokenID,
		session.TokenHash,
		session.IssuedAt,
		session.ExpiresAt,
		session.RevokedAt,
		session.ReplacedByTokenID,
		session.LastUsedAt,
	).Scan(&session.ID, &session.CreatedAt)
}

func (r *RefreshTokenRepository) RotateAndRevoke(
	ctx context.Context,
	tokenID string,
	tokenHash string,
	userID int64,
	usedAt time.Time,
	replacedByTokenID string,
) (bool, error) {
	const query = `
UPDATE auth_refresh_token
SET revoked_at = $4,
	replaced_by_token_id = $5,
	last_used_at = $4
WHERE token_id = $1
  AND token_hash = $2
  AND user_id = $3
  AND revoked_at IS NULL
  AND expires_at > $4`

	cmd, err := r.db.Exec(ctx, query, tokenID, tokenHash, userID, usedAt, replacedByTokenID)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() == 1, nil
}

func (r *RefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenID string,
	tokenHash string,
	userID int64,
	usedAt time.Time,
) (bool, error) {
	const query = `
UPDATE auth_refresh_token
SET revoked_at = $4,
	last_used_at = $4
WHERE token_id = $1
  AND token_hash = $2
  AND user_id = $3
  AND revoked_at IS NULL`

	cmd, err := r.db.Exec(ctx, query, tokenID, tokenHash, userID, usedAt)
	if err != nil {
		return false, err
	}
	return cmd.RowsAffected() == 1, nil
}
