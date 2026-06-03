package http

import (
	authService "CourseJob/internal/service/auth"
	"CourseJob/internal/transport/http/dto"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"
)

// RegisterStudent registers a student user account and returns JWT token.
// @Summary Register student
// @Description Registers a student user and creates both student profile and user account in one transaction.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterStudentRequest true "Student registration payload"
// @Success 201 {object} dto.AuthTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register/student [post]
func (h *Handler) RegisterStudent(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	var req dto.RegisterStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{
			Status: "error",
			Error:  "invalid request body",
		})
		return
	}

	token, err := h.authService.RegisterStudent(r.Context(), authService.RegisterStudentInput{
		FullName: req.FullName,
		Email:    req.Email,
		Password: req.Password,
		Course:   req.Course,
		Group:    req.Group,
		UID:      req.UID,
	})
	if err != nil {
		switch {
		case errors.Is(err, authService.ErrInvalidFullName),
			errors.Is(err, authService.ErrInvalidCourse),
			errors.Is(err, authService.ErrInvalidGroup),
			errors.Is(err, authService.ErrInvalidEmail),
			errors.Is(err, authService.ErrInvalidUID),
			errors.Is(err, authService.ErrWeakPassword):
			writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: err.Error()})
		case errors.Is(err, authService.ErrUserExists),
			errors.Is(err, authService.ErrStudentExists):
			writeJSON(w, nethttp.StatusConflict, ErrorResponse{Status: "error", Error: err.Error()})
		default:
			writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		}
		return
	}

	writeJSON(w, nethttp.StatusCreated, toAuthTokenResponse(token))
}

// Login authenticates user and returns JWT token.
// @Summary Login
// @Description Authenticates by email and password and returns access token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login payload"
// @Success 200 {object} dto.AuthTokenResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{
			Status: "error",
			Error:  "invalid request body",
		})
		return
	}

	token, err := h.authService.Login(r.Context(), authService.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, authService.ErrInvalidCredentials) {
			writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{Status: "error", Error: err.Error()})
			return
		}
		writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		return
	}

	writeJSON(w, nethttp.StatusOK, toAuthTokenResponse(token))
}

// Refresh exchanges refresh token for a new token pair.
// @Summary Refresh token
// @Description Issues a new access/refresh token pair using a valid refresh token.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RefreshRequest true "Refresh token payload"
// @Success 200 {object} dto.AuthTokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	var req dto.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{
			Status: "error",
			Error:  "invalid request body",
		})
		return
	}

	token, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, authService.ErrInvalidRefresh) {
			writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{Status: "error", Error: err.Error()})
			return
		}
		writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		return
	}

	writeJSON(w, nethttp.StatusOK, toAuthTokenResponse(token))
}

// Revoke invalidates a refresh token.
// @Summary Revoke token
// @Description Revokes a refresh token so it can no longer be used.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.RevokeRequest true "Revoke token payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/revoke [post]
func (h *Handler) Revoke(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	var req dto.RevokeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{
			Status: "error",
			Error:  "invalid request body",
		})
		return
	}

	if err := h.authService.Revoke(r.Context(), req.RefreshToken); err != nil {
		if errors.Is(err, authService.ErrInvalidRefresh) {
			writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{Status: "error", Error: err.Error()})
			return
		}
		writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		return
	}

	writeJSON(w, nethttp.StatusOK, map[string]string{
		"status": "success",
	})
}

// CreateUserByAdmin creates teacher/admin user account.
// @Summary Create user by admin
// @Description Creates teacher or admin account. Requires Bearer token with admin role.
// @Tags Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param request body dto.AdminCreateUserRequest true "Create user payload"
// @Success 201 {object} dto.CreatedUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users [post]
func (h *Handler) CreateUserByAdmin(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPost {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	var req dto.AdminCreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{
			Status: "error",
			Error:  "invalid request body",
		})
		return
	}

	created, err := h.authService.CreateUserByAdmin(r.Context(), authService.AdminCreateUserInput{
		Email:    req.Email,
		Password: req.Password,
		Role:     req.Role,
	})
	if err != nil {
		switch {
		case errors.Is(err, authService.ErrInvalidEmail),
			errors.Is(err, authService.ErrWeakPassword),
			errors.Is(err, authService.ErrInvalidRole):
			writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: err.Error()})
		case errors.Is(err, authService.ErrUserExists):
			writeJSON(w, nethttp.StatusConflict, ErrorResponse{Status: "error", Error: err.Error()})
		default:
			writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		}
		return
	}

	writeJSON(w, nethttp.StatusCreated, dto.CreatedUserResponse{
		ID:    created.ID,
		Email: created.Email,
		Role:  created.Role,
	})
}

// ListUsersByAdmin returns users list with optional filters.
// @Summary List users by admin
// @Description Returns users with optional role/email filters. Requires Bearer token with admin role.
// @Tags Admin
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param role query string false "Role filter: student|teacher|admin"
// @Param email query string false "Email contains filter"
// @Param limit query int false "Limit (default 100, max 200)"
// @Param offset query int false "Offset (default 0)"
// @Success 200 {object} dto.AdminUsersListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users [get]
func (h *Handler) ListUsersByAdmin(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodGet {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	limit, offset, err := parseListUsersPagination(r)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: err.Error()})
		return
	}

	users, err := h.authService.ListUsersByAdmin(r.Context(), authService.AdminUsersListFilter{
		Role:       strings.TrimSpace(r.URL.Query().Get("role")),
		EmailQuery: strings.TrimSpace(r.URL.Query().Get("email")),
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		if errors.Is(err, authService.ErrInvalidRole) {
			writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: err.Error()})
			return
		}
		writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		return
	}

	responseUsers := make([]dto.AdminUserResponse, 0, len(users))
	for _, user := range users {
		responseUsers = append(responseUsers, dto.NewAdminUserResponse(
			user.ID,
			user.Email,
			user.Role,
			user.IsActive,
			user.CreatedAt,
		))
	}

	writeJSON(w, nethttp.StatusOK, dto.AdminUsersListResponse{
		Status: "success",
		Users:  responseUsers,
		Count:  len(responseUsers),
	})
}

// UpdateUserRoleByAdmin changes a user role.
// @Summary Update user role
// @Description Updates user role. Allowed roles: teacher or admin. Requires Bearer token with admin role.
// @Tags Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param id path int true "User ID"
// @Param request body dto.AdminUpdateUserRoleRequest true "Role update payload"
// @Success 200 {object} dto.AdminUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users/{id}/role [patch]
func (h *Handler) UpdateUserRoleByAdmin(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPatch {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	userID, err := parseUserIDParam(r)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: err.Error()})
		return
	}

	var req dto.AdminUpdateUserRoleRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: "invalid request body"})
		return
	}

	updated, err := h.authService.UpdateUserRoleByAdmin(r.Context(), userID, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, authService.ErrInvalidRole):
			writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: err.Error()})
		case errors.Is(err, authService.ErrUserNotFound):
			writeJSON(w, nethttp.StatusNotFound, ErrorResponse{Status: "error", Error: err.Error()})
		default:
			writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		}
		return
	}

	writeJSON(w, nethttp.StatusOK, dto.NewAdminUserResponse(
		updated.ID,
		updated.Email,
		updated.Role,
		updated.IsActive,
		updated.CreatedAt,
	))
}

// UpdateUserActiveByAdmin activates/deactivates a user.
// @Summary Update user active flag
// @Description Updates user is_active status. Requires Bearer token with admin role.
// @Tags Admin
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer access token"
// @Param id path int true "User ID"
// @Param request body dto.AdminUpdateUserActiveRequest true "Active update payload"
// @Success 200 {object} dto.AdminUserResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 403 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /admin/users/{id}/active [patch]
func (h *Handler) UpdateUserActiveByAdmin(w nethttp.ResponseWriter, r *nethttp.Request) {
	if r.Method != nethttp.MethodPatch {
		writeJSON(w, nethttp.StatusMethodNotAllowed, ErrorResponse{
			Status: "error",
			Error:  "method not allowed",
		})
		return
	}

	userID, err := parseUserIDParam(r)
	if err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: err.Error()})
		return
	}

	var req dto.AdminUpdateUserActiveRequest
	if err = json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, nethttp.StatusBadRequest, ErrorResponse{Status: "error", Error: "invalid request body"})
		return
	}

	updated, err := h.authService.UpdateUserActiveByAdmin(r.Context(), userID, req.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, authService.ErrUserNotFound):
			writeJSON(w, nethttp.StatusNotFound, ErrorResponse{Status: "error", Error: err.Error()})
		default:
			writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{Status: "error", Error: "internal server error"})
		}
		return
	}

	writeJSON(w, nethttp.StatusOK, dto.NewAdminUserResponse(
		updated.ID,
		updated.Email,
		updated.Role,
		updated.IsActive,
		updated.CreatedAt,
	))
}

func parseUserIDParam(r *nethttp.Request) (int64, error) {
	raw := strings.TrimSpace(chi.URLParam(r, "id"))
	if raw == "" {
		return 0, errors.New("user id is required")
	}

	userID, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || userID <= 0 {
		return 0, errors.New("invalid user id")
	}

	return userID, nil
}

func parseListUsersPagination(r *nethttp.Request) (int, int, error) {
	limit := 100
	offset := 0

	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		value, err := strconv.Atoi(rawLimit)
		if err != nil || value <= 0 || value > 200 {
			return 0, 0, errors.New("limit must be an integer from 1 to 200")
		}
		limit = value
	}

	if rawOffset := strings.TrimSpace(r.URL.Query().Get("offset")); rawOffset != "" {
		value, err := strconv.Atoi(rawOffset)
		if err != nil || value < 0 {
			return 0, 0, errors.New("offset must be a non-negative integer")
		}
		offset = value
	}

	return limit, offset, nil
}

func toAuthTokenResponse(token *authService.AuthToken) dto.AuthTokenResponse {
	return dto.AuthTokenResponse{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		TokenType:        token.TokenType,
		ExpiresAt:        token.ExpiresAt.UTC().Format(time.RFC3339),
		ExpiresIn:        token.ExpiresIn,
		RefreshExpiresAt: token.RefreshExpiresAt.UTC().Format(time.RFC3339),
		RefreshExpiresIn: token.RefreshExpiresIn,
		UserID:           token.UserID,
		Email:            token.Email,
		Role:             token.Role,
	}
}
