package auth

import (
	authA "CourseJob/internal/auth"
	"CourseJob/internal/domain"
	"CourseJob/internal/storage/postgres"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrUserExists         = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrStudentExists      = errors.New("student with this email or uid already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidRefresh     = errors.New("invalid refresh token")
	ErrInvalidRole        = errors.New("invalid role")
	ErrInvalidUID         = errors.New("invalid uid")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidFullName    = errors.New("full name is required")
	ErrInvalidGroup       = errors.New("group is required")
	ErrInvalidCourse      = errors.New("course must be between 1 and 4")
	ErrWeakPassword       = errors.New("password must be at least 8 characters")
)

type Service struct {
	transactor     postgres.Transactor
	uidHasher      authA.Hasher
	passwordHasher authA.PasswordHasher
	tokenManager   authA.TokenManager
}

type AuthToken struct {
	AccessToken      string
	RefreshToken     string
	TokenType        string
	ExpiresAt        time.Time
	ExpiresIn        int64
	RefreshExpiresAt time.Time
	RefreshExpiresIn int64
	UserID           int64
	Email            string
	Role             string
}

type RegisterStudentInput struct {
	FullName string
	Email    string
	Password string
	Course   int
	Group    string
	UID      string
}

type LoginInput struct {
	Email    string
	Password string
}

type AdminCreateUserInput struct {
	Email    string
	Password string
	Role     string
}

type CreatedUser struct {
	ID    int64
	Email string
	Role  string
}

type AdminUserSummary struct {
	ID        int64
	Email     string
	Role      string
	IsActive  bool
	CreatedAt time.Time
}

type AdminUsersListFilter struct {
	Role       string
	EmailQuery string
	Limit      int
	Offset     int
}

type Principal struct {
	UserID int64
	Email  string
	Role   string
}

func NewService(
	transactor postgres.Transactor,
	uidHasher authA.Hasher,
	passwordHasher authA.PasswordHasher,
	tokenManager authA.TokenManager,
) *Service {
	return &Service{
		transactor:     transactor,
		uidHasher:      uidHasher,
		passwordHasher: passwordHasher,
		tokenManager:   tokenManager,
	}
}

func (s *Service) RegisterStudent(ctx context.Context, input RegisterStudentInput) (*AuthToken, error) {
	fullName := strings.TrimSpace(input.FullName)
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	groupName := strings.TrimSpace(input.Group)
	course := input.Course
	uid, ok := authA.NormalizeAndValidateCardUID(input.UID)

	if fullName == "" {
		return nil, ErrInvalidFullName
	}
	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if course < 1 || course > 4 {
		return nil, ErrInvalidCourse
	}
	if groupName == "" {
		return nil, ErrInvalidGroup
	}
	if !ok {
		return nil, ErrInvalidUID
	}
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, err
	}

	uidHash, err := s.uidHasher.Hash(uid)
	if err != nil {
		return nil, err
	}

	var createdUser domain.User
	createdUser.Email = email
	createdUser.PasswordHash = passwordHash
	createdUser.Role = domain.RoleStudent
	createdUser.IsActive = true

	err = s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		student := &domain.Student{
			FullName:  fullName,
			Course:    course,
			GroupName: groupName,
			Email:     email,
			CardUID:   uidHash,
			CreatedAt: time.Now().UTC(),
		}

		if createStudentErr := repo.Students().CreateStudent(ctx, student); createStudentErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(createStudentErr, &pgErr) && pgErr.Code == "23505" {
				return ErrStudentExists
			}
			return createStudentErr
		}

		createdStudent, lookupErr := repo.Students().GetByCardUID(ctx, uidHash)
		if lookupErr != nil {
			return lookupErr
		}
		if createdStudent == nil {
			return errors.New("failed to load created student")
		}

		createdUser.StudentID = &createdStudent.ID
		if createErr := repo.Users().Create(ctx, &createdUser); createErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(createErr, &pgErr) && pgErr.Code == "23505" {
				return ErrUserExists
			}
			return createErr
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return s.issueToken(ctx, createdUser)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (*AuthToken, error) {
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	if !isValidEmail(email) {
		return nil, ErrInvalidCredentials
	}
	if password == "" {
		return nil, ErrInvalidCredentials
	}

	var user *domain.User
	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		found, findErr := repo.Users().GetByEmail(ctx, email)
		if findErr != nil {
			return findErr
		}
		user = found
		return nil
	})
	if err != nil {
		return nil, err
	}

	if user == nil || !user.IsActive {
		return nil, ErrInvalidCredentials
	}
	if !s.passwordHasher.Verify(password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
	}

	return s.issueToken(ctx, *user)
}

func (s *Service) CreateUserByAdmin(ctx context.Context, input AdminCreateUserInput) (*CreatedUser, error) {
	email := normalizeEmail(input.Email)
	password := strings.TrimSpace(input.Password)
	role := strings.ToLower(strings.TrimSpace(input.Role))

	if !isValidEmail(email) {
		return nil, ErrInvalidEmail
	}
	if len(password) < 8 {
		return nil, ErrWeakPassword
	}
	if role != domain.RoleTeacher && role != domain.RoleAdmin {
		return nil, fmt.Errorf("%w: role must be teacher or admin", ErrInvalidRole)
	}

	passwordHash, err := s.passwordHasher.Hash(password)
	if err != nil {
		return nil, err
	}

	user := domain.User{
		Email:        email,
		PasswordHash: passwordHash,
		Role:         role,
		IsActive:     true,
	}

	err = s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		if createErr := repo.Users().Create(ctx, &user); createErr != nil {
			var pgErr *pgconn.PgError
			if errors.As(createErr, &pgErr) && pgErr.Code == "23505" {
				return ErrUserExists
			}
			return createErr
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &CreatedUser{
		ID:    user.ID,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *Service) ListUsersByAdmin(ctx context.Context, filter AdminUsersListFilter) ([]AdminUserSummary, error) {
	role := strings.ToLower(strings.TrimSpace(filter.Role))
	if role != "" && role != domain.RoleStudent && role != domain.RoleTeacher && role != domain.RoleAdmin {
		return nil, fmt.Errorf("%w: role must be student, teacher or admin", ErrInvalidRole)
	}

	users := make([]domain.User, 0)
	err := s.transactor.WithoutTransaction(ctx, func(repo postgres.Repository) error {
		var repoErr error
		users, repoErr = repo.Users().List(ctx, role, filter.EmailQuery, filter.Limit, filter.Offset)
		return repoErr
	})
	if err != nil {
		return nil, err
	}

	result := make([]AdminUserSummary, 0, len(users))
	for _, user := range users {
		result = append(result, AdminUserSummary{
			ID:        user.ID,
			Email:     user.Email,
			Role:      user.Role,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
		})
	}

	return result, nil
}

func (s *Service) UpdateUserRoleByAdmin(ctx context.Context, userID int64, role string) (*AdminUserSummary, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	role = strings.ToLower(strings.TrimSpace(role))
	if role != domain.RoleTeacher && role != domain.RoleAdmin {
		return nil, fmt.Errorf("%w: role must be teacher or admin", ErrInvalidRole)
	}

	var updated *domain.User
	err := s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		var repoErr error
		updated, repoErr = repo.Users().UpdateRole(ctx, userID, role)
		if repoErr != nil {
			return repoErr
		}
		if updated == nil {
			return ErrUserNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AdminUserSummary{
		ID:        updated.ID,
		Email:     updated.Email,
		Role:      updated.Role,
		IsActive:  updated.IsActive,
		CreatedAt: updated.CreatedAt,
	}, nil
}

func (s *Service) UpdateUserActiveByAdmin(ctx context.Context, userID int64, isActive bool) (*AdminUserSummary, error) {
	if userID <= 0 {
		return nil, ErrUserNotFound
	}

	var updated *domain.User
	err := s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		var repoErr error
		updated, repoErr = repo.Users().UpdateActive(ctx, userID, isActive)
		if repoErr != nil {
			return repoErr
		}
		if updated == nil {
			return ErrUserNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &AdminUserSummary{
		ID:        updated.ID,
		Email:     updated.Email,
		Role:      updated.Role,
		IsActive:  updated.IsActive,
		CreatedAt: updated.CreatedAt,
	}, nil
}

func (s *Service) ParseAccessToken(token string) (*Principal, error) {
	claims, err := s.tokenManager.ParseAccess(token)
	if err != nil {
		return nil, err
	}

	return &Principal{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (*AuthToken, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, ErrInvalidRefresh
	}

	claims, err := s.tokenManager.ParseRefresh(refreshToken)
	if err != nil {
		return nil, ErrInvalidRefresh
	}

	oldTokenID := strings.TrimSpace(claims.ID)
	if oldTokenID == "" {
		return nil, ErrInvalidRefresh
	}
	oldTokenHash := hashRefreshToken(refreshToken)

	var issuedToken *AuthToken
	err = s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		user, findErr := repo.Users().GetByID(ctx, claims.UserID)
		if findErr != nil {
			return findErr
		}
		if user == nil || !user.IsActive {
			return ErrInvalidRefresh
		}
		if user.Role != claims.Role || normalizeEmail(user.Email) != normalizeEmail(claims.Email) {
			return ErrInvalidRefresh
		}

		accessToken, accessExpiresAt, genAccessErr := s.tokenManager.GenerateAccess(user.ID, user.Email, user.Role)
		if genAccessErr != nil {
			return genAccessErr
		}
		newRefreshToken, newRefreshTokenID, newRefreshExpiresAt, genRefreshErr := s.tokenManager.GenerateRefresh(user.ID, user.Email, user.Role)
		if genRefreshErr != nil {
			return genRefreshErr
		}

		now := time.Now().UTC()
		rotated, rotateErr := repo.RefreshTokens().RotateAndRevoke(
			ctx,
			oldTokenID,
			oldTokenHash,
			user.ID,
			now,
			newRefreshTokenID,
		)
		if rotateErr != nil {
			return rotateErr
		}
		if !rotated {
			return ErrInvalidRefresh
		}

		newSession := &domain.RefreshTokenSession{
			UserID:    user.ID,
			TokenID:   newRefreshTokenID,
			TokenHash: hashRefreshToken(newRefreshToken),
			IssuedAt:  now,
			ExpiresAt: newRefreshExpiresAt,
		}
		if createErr := repo.RefreshTokens().Create(ctx, newSession); createErr != nil {
			return createErr
		}

		issuedToken = composeAuthToken(*user, accessToken, accessExpiresAt, newRefreshToken, newRefreshExpiresAt)
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidRefresh) {
			return nil, ErrInvalidRefresh
		}
		return nil, err
	}

	return issuedToken, nil
}

func (s *Service) Revoke(ctx context.Context, refreshToken string) error {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return ErrInvalidRefresh
	}

	claims, err := s.tokenManager.ParseRefresh(refreshToken)
	if err != nil {
		return ErrInvalidRefresh
	}
	tokenID := strings.TrimSpace(claims.ID)
	if tokenID == "" {
		return ErrInvalidRefresh
	}
	tokenHash := hashRefreshToken(refreshToken)
	now := time.Now().UTC()

	err = s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		revoked, revokeErr := repo.RefreshTokens().Revoke(ctx, tokenID, tokenHash, claims.UserID, now)
		if revokeErr != nil {
			return revokeErr
		}
		if !revoked {
			return ErrInvalidRefresh
		}
		return nil
	})
	if err != nil {
		if errors.Is(err, ErrInvalidRefresh) {
			return ErrInvalidRefresh
		}
		return err
	}
	return nil
}

func (s *Service) issueToken(ctx context.Context, user domain.User) (*AuthToken, error) {
	accessToken, accessExpiresAt, err := s.tokenManager.GenerateAccess(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenID, refreshExpiresAt, err := s.tokenManager.GenerateRefresh(user.ID, user.Email, user.Role)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	err = s.transactor.WithinTransaction(ctx, func(repo postgres.Repository) error {
		session := &domain.RefreshTokenSession{
			UserID:    user.ID,
			TokenID:   refreshTokenID,
			TokenHash: hashRefreshToken(refreshToken),
			IssuedAt:  now,
			ExpiresAt: refreshExpiresAt,
		}
		return repo.RefreshTokens().Create(ctx, session)
	})
	if err != nil {
		return nil, err
	}

	return composeAuthToken(user, accessToken, accessExpiresAt, refreshToken, refreshExpiresAt), nil
}

func composeAuthToken(user domain.User, accessToken string, accessExpiresAt time.Time, refreshToken string, refreshExpiresAt time.Time) *AuthToken {
	return &AuthToken{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		TokenType:        "Bearer",
		ExpiresAt:        accessExpiresAt,
		ExpiresIn:        int64(time.Until(accessExpiresAt).Seconds()),
		RefreshExpiresAt: refreshExpiresAt,
		RefreshExpiresIn: int64(time.Until(refreshExpiresAt).Seconds()),
		UserID:           user.ID,
		Email:            user.Email,
		Role:             user.Role,
	}
}

func hashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func isValidEmail(email string) bool {
	if email == "" {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return addr.Address == email
}
