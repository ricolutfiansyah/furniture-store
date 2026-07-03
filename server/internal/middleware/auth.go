package middleware

import (
	"context"
	"errors"
	"furniture-api/internal/domain"
	"furniture-api/internal/repository"
	"furniture-api/internal/response"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthenticatedUser struct {
	ID       int
	PublicID string
	Role     string
}

type UserFinder interface {
	FindByPublicID(ctx context.Context, publicID string) (*domain.User, error)
}

func AuthMiddleware(jwtSecret string, userRepo UserFinder) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.WriteError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.WriteError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			token, err := jwt.Parse(parts[1], func(token *jwt.Token) (any, error) {
				return []byte(jwtSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil || !token.Valid {
				response.WriteError(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			publicID, ok := claims["sub"].(string)
			if !ok {
				response.WriteError(w, http.StatusUnauthorized, "invalid user id in token")
				return
			}

			user, err := userRepo.FindByPublicID(r.Context(), publicID)
			if err != nil {
				if errors.Is(err, repository.ErrUserNotFound) {
					response.WriteError(w, http.StatusUnauthorized, "user not found")
					return
				}
				log.Printf("auth middleware find user: %w", err)
				response.WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			if !user.IsActive {
				response.WriteError(w, http.StatusForbidden, "account is inactive")
				return
			}

			authUser := AuthenticatedUser{
				ID:       user.ID,
				PublicID: user.PublicID,
				Role:     user.Role,
			}

			ctx := context.WithValue(r.Context(), UserContextKey, authUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(ctx context.Context) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(UserContextKey).(AuthenticatedUser)
	return user, ok
}

func AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, ok := GetUserFromContext(r.Context())
		if !ok {
			response.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if user.Role != "admin" {
			response.WriteError(w, http.StatusForbidden, "admin access required")
			return
		}

		next.ServeHTTP(w, r)
	})
}
