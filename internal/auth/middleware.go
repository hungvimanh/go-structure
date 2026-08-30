package auth

import (
	"crypto/rsa"
	"log"
	"net/http"
	"strings"

	"category-service/internal/shared/httpresponse"
	"category-service/internal/shared/usercontext"
)

const bearerPrefix = "Bearer "

func Middleware(publicKey *rsa.PublicKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, err := extractBearerToken(r)
			if err != nil {
				log.Printf("unauthorized: %v", err)
				httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := VerifyToken(token, publicKey)
			if err != nil {
				log.Printf("unauthorized: %v", err)
				httpresponse.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := usercontext.WithUser(r.Context(), newUserContext(claims))
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// newUserContext anh xa Claims (JWT tho) sang UserContext (du lieu nghiep vu).
// Claims co them field moi thi bo sung anh xa tuong ung vao day.
func newUserContext(claims *Claims) *usercontext.UserContext {
	return &usercontext.UserContext{
		UserID:   claims.UserID,
		Email:    claims.Email,
		FullName: claims.FullName,
	}
}

func extractBearerToken(r *http.Request) (string, error) {
	header := r.Header.Get("Authorization")
	if header == "" || !strings.HasPrefix(header, bearerPrefix) {
		return "", ErrInvalidToken
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix))
	if token == "" {
		return "", ErrInvalidToken
	}

	return token, nil
}
