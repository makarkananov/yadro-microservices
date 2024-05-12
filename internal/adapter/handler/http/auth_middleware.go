package http

import (
	"context"
	"log"
	"net/http"
	"yadro-microservices/internal/core/domain"
	"yadro-microservices/internal/core/port"
)

type key int

const currentUserKey key = 0

// AuthorizationMiddleware is a middleware that checks if the user is authorized to access the resource.
func AuthorizationMiddleware(role domain.Role) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			userData := r.Context().Value(currentUserKey)
			user, _ := userData.(*domain.User)

			if role != domain.UNDEFINED && (user == nil || user.Role != domain.ADMIN) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}

// AuthenticationMiddleware is a middleware that checks if the user is
// authenticated. If required is true, the middleware will return an error if the
// user is not authenticated.
func AuthenticationMiddleware(authService port.AuthService, required bool) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			tokenString := r.Header.Get("Authorization")
			if len(tokenString) <= len("Bearer ") {
				if required {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			user, err := authService.ValidateToken(r.Context(), tokenString[len("Bearer "):])
			if err != nil {
				log.Printf("Error validating token: %v", err)
				if required {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}

				next.ServeHTTP(w, r)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), currentUserKey, user)))
		}
	}
}