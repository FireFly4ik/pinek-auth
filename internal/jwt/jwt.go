package jwt

import (
	"auth/internal/config"
	"crypto/rsa"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type JWTService interface {
	GenerateAccessToken(userId, username, role string) (string, error)
	ValidateAccessToken(tokenStr string) (*AccessClaims, error)
	GenerateRefreshToken(userId, sessionId string) (string, error)
	ValidateRefreshToken(tokenStr string) (*RefreshClaims, error)
}

type jwtService struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	envConf    *config.Config
}

func NewJWTService(envConf *config.Config) JWTService {
	pubKey, err := LoadRSAPublicKey()
	if err != nil {
		panic("not existed public key file")
	}

	priKey, err := LoadRSAPrivateKey()
	if err != nil {
		panic("not existed private key file")
	}

	return &jwtService{
		privateKey: priKey,
		publicKey:  pubKey,
		envConf:    envConf,
	}
}

func (s *jwtService) GenerateAccessToken(userID, username, role string) (string, error) {
	accessTTL, err := time.ParseDuration(s.envConf.JWT.AccessTTL)
	if err != nil {
		return "", err
	}

	claims := AccessClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *jwtService) ValidateAccessToken(tokenStr string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &AccessClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}

func (s *jwtService) GenerateRefreshToken(userID, sessionID string) (string, error) {
	refreshTTL, err := time.ParseDuration(s.envConf.JWT.RefreshTTL)
	if err != nil {
		return "", err
	}

	claims := RefreshClaims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(s.privateKey)
}

func (s *jwtService) ValidateRefreshToken(tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
