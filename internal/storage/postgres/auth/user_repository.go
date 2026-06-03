package auth

import (
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres/contract"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

type UserRepository struct {
	db contract.DBTX
}

func NewUserRepository(db contract.DBTX) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	const query = `
INSERT INTO users (email, password_hash, role, student_id, teacher_id, is_active)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, created_at`

	return r.db.QueryRow(
		ctx,
		query,
		user.Email,
		user.PasswordHash,
		user.Role,
		user.StudentID,
		user.TeacherID,
		user.IsActive,
	).Scan(&user.ID, &user.CreatedAt)
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
SELECT id, email, password_hash, role, student_id, teacher_id, is_active, created_at
FROM users
WHERE email = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.StudentID,
		&user.TeacherID,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	const query = `
SELECT id, email, password_hash, role, student_id, teacher_id, is_active, created_at
FROM users
WHERE id = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.StudentID,
		&user.TeacherID,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	const query = `
SELECT COUNT(*)
FROM users
WHERE role = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, role).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (r *UserRepository) List(ctx context.Context, role string, emailQuery string, limit int, offset int) ([]domain.User, error) {
	baseQuery := `
SELECT id, email, password_hash, role, student_id, teacher_id, is_active, created_at
FROM users
WHERE 1=1`

	args := make([]any, 0, 4)
	role = strings.ToLower(strings.TrimSpace(role))
	if role != "" {
		args = append(args, role)
		baseQuery += fmt.Sprintf(" AND role = $%d", len(args))
	}

	emailQuery = strings.ToLower(strings.TrimSpace(emailQuery))
	if emailQuery != "" {
		args = append(args, "%"+emailQuery+"%")
		baseQuery += fmt.Sprintf(" AND lower(email) LIKE $%d", len(args))
	}

	if limit <= 0 || limit > 200 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	args = append(args, limit, offset)
	baseQuery += fmt.Sprintf(" ORDER BY created_at DESC, id DESC LIMIT $%d OFFSET $%d", len(args)-1, len(args))

	rows, err := r.db.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]domain.User, 0, limit)
	for rows.Next() {
		var user domain.User
		if err = rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.StudentID,
			&user.TeacherID,
			&user.IsActive,
			&user.CreatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, id int64, role string) (*domain.User, error) {
	const query = `
UPDATE users
SET role = $2
WHERE id = $1
RETURNING id, email, password_hash, role, student_id, teacher_id, is_active, created_at`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id, role).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.StudentID,
		&user.TeacherID,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) UpdateActive(ctx context.Context, id int64, isActive bool) (*domain.User, error) {
	const query = `
UPDATE users
SET is_active = $2
WHERE id = $1
RETURNING id, email, password_hash, role, student_id, teacher_id, is_active, created_at`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id, isActive).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.StudentID,
		&user.TeacherID,
		&user.IsActive,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}
