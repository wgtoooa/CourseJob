package http

import (
	"net/http"
	"strings"
)

func CORS() func(next http.Handler) http.Handler {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173":           {},
		"http://localhost:4173":           {},
		"https://yaroslavka123.github.io": {},
		"https://rfict.up.railway.app":    {},
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin != "" {
				if _, ok := allowedOrigins[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Set("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
