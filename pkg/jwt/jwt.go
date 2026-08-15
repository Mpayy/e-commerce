package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/Mpayy/e-commerce/pkg/apperror"
	"github.com/Mpayy/e-commerce/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

const TokenDuration = 24 * time.Hour * 30

type Auth struct {
	ID   uint
	Role string
}

type CustomClaim struct {
	ID   uint   `json:"id"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

//go:generate mockery

//mockery:generate: true
//mockery:filename: ../../internal/user/mocks/mock_jwt.go
type JwtToken interface {
	Create(auth *Auth) (string, error)
	Validate(jwtToken string) (*Auth, error)
}

type JwtTokenImpl struct {
	SecretKey string
}

func NewJwtToken(cfg *config.Config) JwtToken {
	secretKey := cfg.JWTSecretKey
	return &JwtTokenImpl{
		SecretKey: secretKey,
	}
}

func (t *JwtTokenImpl) Create(auth *Auth) (string, error) {
	claims := CustomClaim{
		ID:   auth.ID,
		Role: auth.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtToken, err := token.SignedString([]byte(t.SecretKey))
	if err != nil {
		return "", fmt.Errorf("create token failed: %w", err)
	}

	return jwtToken, nil
}

func (t *JwtTokenImpl) Validate(jwtToken string) (*Auth, error) {
	var claims CustomClaim
	token, err := jwt.ParseWithClaims(jwtToken, &claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, apperror.ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) || errors.Is(err, jwt.ErrTokenMalformed) {
			return nil, apperror.ErrInvalidToken
		}
		return nil, err
	}

	if !token.Valid {
		return nil, apperror.ErrInvalidToken
	}

	return &Auth{
		ID: claims.ID,
		Role: claims.Role,
	}, nil
}
