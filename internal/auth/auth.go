package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ksusonic/gophermart/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	defaultJwtKey    = "my_secret_key"
	userIDContextKey = "user_id"
)

type Controller struct {
	jwtKey []byte
}

func NewAuthController(jwtKey string) *Controller {
	key := defaultJwtKey
	if jwtKey != "" {
		key = jwtKey
	}

	return &Controller{
		jwtKey: []byte(key),
	}
}

func (c *Controller) IsAuthorized() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Cookie("Authorization")
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := c.parseToken(cookie)
		if err != nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set(userIDContextKey, claims.UserID)
		ctx.Next()
	}
}

func (c *Controller) GetUserID(ctx *gin.Context) (uint, error) {
	userID, ok := ctx.Get(userIDContextKey)
	if !ok {
		return 0, fmt.Errorf("user_id not found in context")
	}
	return userID.(uint), nil
}

func (c *Controller) CreateSignedJWT(claims models.Claims, expiresAt time.Time) (string, error) {
	claims.RegisteredClaims = jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		Issuer:    "gophermart server",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(c.jwtKey)
}

func (c *Controller) parseToken(tokenString string) (claims *models.Claims, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return c.jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*models.Claims)

	if !ok {
		return nil, err
	}

	return claims, nil
}
