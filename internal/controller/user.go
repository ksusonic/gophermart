package controller

import (
	"net/http"
	"time"

	"github.com/ksusonic/gophermart/internal/database"
	"github.com/ksusonic/gophermart/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	Controller
	host     string
	jwtToken []byte
}

func NewUserController(host string, db *database.DB, logger *zap.SugaredLogger) *UserController {
	return &UserController{
		Controller: Controller{
			DB:     db,
			Logger: logger,
		},
		host:     host,
		jwtToken: getJwtToken(),
	}
}

func (c *UserController) RegisterHandlers(router *gin.RouterGroup) {
	router.POST("/register", c.registerHandler)
	router.POST("/login", c.loginHandler)
	router.POST("/orders", c.ordersPostHandler)

	router.GET("/orders", c.ordersGetHandler)
	router.GET("/balance", c.balanceHandler)
	router.GET("/balance/withdraw", c.balanceWithdrawHandler)
	router.GET("/withdrawals", c.withdrawalsHandler)
}

func (c *UserController) registerHandler(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User

	c.DB.Orm.Where("login = ?", user.Login).First(&existingUser)

	if existingUser.ID != 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	var err error
	user.Password, err = hashPassword(user.Password, c.Logger)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not generate password hash"})
		return
	}

	c.DB.Orm.Create(&user)
	ctx.JSON(http.StatusOK, gin.H{"status": "created"})
}

func (c *UserController) loginHandler(ctx *gin.Context) {
	var user models.User

	if err := ctx.ShouldBindJSON(&user); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	c.DB.Orm.Where("login = ?", user.Login).First(&existingUser)
	if existingUser.ID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
		return
	}

	errHash := compareHash(user.Password, existingUser.Password)

	if !errHash {
		ctx.JSON(400, gin.H{"error": "invalid password"})
		return
	}

	expiresAt := time.Now().Add(120 * time.Minute)
	signedToken, err := createJWT(user.Login, expiresAt)
	if err != nil {
		ctx.JSON(500, gin.H{"error": "could not generate token"})
		return
	}

	ctx.SetCookie("token", signedToken, int(expiresAt.Unix()), "/", c.host, false, true)
	ctx.JSON(200, gin.H{"success": "logged in"})
}

func (c *UserController) ordersPostHandler(ctx *gin.Context) {
	// auth-only
}

func (c *UserController) ordersGetHandler(ctx *gin.Context) {
	// auth-only
}

func (c *UserController) balanceHandler(ctx *gin.Context) {
	// auth-only
}

func (c *UserController) balanceWithdrawHandler(ctx *gin.Context) {
	// auth-only
}

func (c *UserController) withdrawalsHandler(ctx *gin.Context) {
	// auth-only
}
