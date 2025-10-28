package jwt

import "github.com/golang-jwt/jwt/v5"

type AccessClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role,omitempty"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	jwt.RegisteredClaims
}
