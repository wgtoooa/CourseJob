package http

import (
	authService "CourseJob/internal/service/auth"
	"context"
	nethttp "net/http"
	"strings"
)

type principalContextKey string

const authPrincipalKey principalContextKey = "auth_principal"

func (h *Handler) AuthRequired(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		if h.authService == nil {
			writeJSON(w, nethttp.StatusInternalServerError, ErrorResponse{
				Status: "error",
				Error:  "auth service is not configured",
			})
			return
		}

		rawHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if rawHeader == "" {
			writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{
				Status: "error",
				Error:  "authorization header is required",
			})
			return
		}

		parts := strings.SplitN(rawHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{
				Status: "error",
				Error:  "authorization header must be Bearer token",
			})
			return
		}

		principal, err := h.authService.ParseAccessToken(strings.TrimSpace(parts[1]))
		if err != nil {
			writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{
				Status: "error",
				Error:  "invalid access token",
			})
			return
		}

		ctx := context.WithValue(r.Context(), authPrincipalKey, *principal)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (h *Handler) RequireRoles(roles ...string) func(next nethttp.Handler) nethttp.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[strings.ToLower(strings.TrimSpace(role))] = struct{}{}
	}

	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				writeJSON(w, nethttp.StatusUnauthorized, ErrorResponse{
					Status: "error",
					Error:  "unauthorized",
				})
				return
			}
			if _, isAllowed := allowed[strings.ToLower(principal.Role)]; !isAllowed {
				writeJSON(w, nethttp.StatusForbidden, ErrorResponse{
					Status: "error",
					Error:  "insufficient role",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func PrincipalFromContext(ctx context.Context) (authService.Principal, bool) {
	value := ctx.Value(authPrincipalKey)
	if value == nil {
		return authService.Principal{}, false
	}

	principal, ok := value.(authService.Principal)
	if !ok {
		return authService.Principal{}, false
	}

	return principal, true
}
