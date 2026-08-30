package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

// Claims la noi extract du lieu tho tu payload JWT.
// JWT sau nay co them field nao, bo sung field + json tag vao day truoc.
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	jwt.RegisteredClaims
}
