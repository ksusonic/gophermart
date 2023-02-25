package controller

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ksusonic/gophermart/internal/utils"

	"go.uber.org/zap"
)

var defaultJwtKey = []byte("my_secret_key")

type Claims struct {
	jwt.RegisteredClaims
	Login string `json:"login"`
}

func getJwtToken() []byte {
	if token := os.Getenv("JWT_TOKEN"); token != "" {
		return []byte(token)
	}

	return defaultJwtKey
}

func hashPassword(password string, logger *zap.SugaredLogger) (string, error) {
	hashed, err := utils.GenerateHashPassword(password)
	if err != nil {
		logger.Warnf("could not hash password: %v", err)
		return password, err
	}

	return hashed, nil
}

func compareHash(password, hash string) bool {
	return utils.CompareHashPassword(password, hash)
}

func createJWT(login string, expiresAt time.Time) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			Issuer:    "gophermart server",
		},
		Login: login,
	})

	return token.SigningString()
}
