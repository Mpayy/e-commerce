package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

const TokenDuration = 24 * time.Hour * 30

type Auth struct {
	UserID   uint
	Role string
}

type JwtToken interface {
	CreateToken(auth *Auth) (string, error)
	ParseToken(token string) (*Auth, error)
}

type JwtTokenImpl struct {
	SecretKey string
}

func NewJwtToken(config *viper.Viper) JwtToken {
	secretKey := config.GetString("JWT_SECRET_KEY")
	return &JwtTokenImpl{
		SecretKey: secretKey,
	}
}

func (t *JwtTokenImpl) CreateToken(auth *Auth) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": auth.UserID,
		"role":    auth.Role,
		"exp":     time.Now().Add(TokenDuration).Unix(),
	})

	jwtToken, err := token.SignedString([]byte(t.SecretKey))
	if err != nil {
		return "", err
	}

	return jwtToken, nil
}

func (t *JwtTokenImpl) ParseToken(jwtToken string) (*Auth, error) {
	token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (any, error) {
		return []byte(t.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claim, ok := token.Claims.(jwt.MapClaims)

	if !ok {
		return nil, err
	}

	id, ok := claim["user_id"].(float64)
	if !ok {
		return nil, err
	}

	role, ok := claim["role"].(string)
	if !ok {
		return nil, err
	}

	auth := &Auth{
		UserID: uint(id),
		Role:   role,
	}

	return auth, nil

}