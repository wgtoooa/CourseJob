package dto

import "time"

type RegisterStudentRequest struct {
	FullName string `json:"fullname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Course   int    `json:"course"`
	Group    string `json:"group"`
	UID      string `json:"uid"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RevokeRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AdminCreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

type AuthTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	TokenType        string `json:"token_type"`
	ExpiresAt        string `json:"expires_at"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresAt string `json:"refresh_expires_at"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	UserID           int64  `json:"user_id"`
	Email            string `json:"email"`
	Role             string `json:"role"`
}

type CreatedUserResponse struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AdminUserResponse struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

type AdminUsersListResponse struct {
	Status string              `json:"status"`
	Users  []AdminUserResponse `json:"users"`
	Count  int                 `json:"count"`
}

type AdminUpdateUserRoleRequest struct {
	Role string `json:"role"`
}

type AdminUpdateUserActiveRequest struct {
	IsActive bool `json:"is_active"`
}

func NewAdminUserResponse(id int64, email, role string, isActive bool, createdAt time.Time) AdminUserResponse {
	return AdminUserResponse{
		ID:        id,
		Email:     email,
		Role:      role,
		IsActive:  isActive,
		CreatedAt: createdAt.UTC().Format(time.RFC3339),
	}
}
