package auth

import (
	"crypto/rsa"
	"net/http"
	"strings"

	"category-service/internal/shared/httpresponse"
	"category-service/internal/shared/usercontext"

	"github.com/labstack/echo/v5"
)

const bearerPrefix = "Bearer "

func Middleware(publicKey *rsa.PublicKey) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			token, err := extractBearerToken(req)
			if err != nil {
				return httpresponse.NewAuthenticationError(err)
			}

			claims, err := VerifyToken(token, publicKey)
			if err != nil {
				return httpresponse.NewAuthenticationError(err)
			}

			ctx := usercontext.WithUser(req.Context(), newUserContext(claims))
			c.SetRequest(req.WithContext(ctx))
			return next(c)
		}
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
