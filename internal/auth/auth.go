package auth

import (
	"time"

	"github.com/ksusonic/gophermart/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

var defaultJwtKey = []byte("my_secret_key")

type Controller struct {
	jwtKey []byte
}

func NewAuthController(jwtKey string) *Controller {
	if jwtKey != "" {
		return &Controller{jwtKey: []byte(jwtKey)}
	} else {
		return &Controller{jwtKey: defaultJwtKey}
	}
}

func (c *Controller) IsAuthorized() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cookie, err := ctx.Cookie("token")

		if err != nil {
			ctx.JSON(401, gin.H{"error": "unauthorized"})
			ctx.Abort()
			return
		}

		claims, err := c.parseToken(cookie)

		if err != nil {
			ctx.JSON(401, gin.H{"error": "unauthorized"})
			ctx.Abort()
			return
		}

		ctx.Set("login", claims.Login)
		ctx.Next()
	}
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
