package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	Controller
}

func NewUserController(logger *zap.SugaredLogger) *UserController {
	return &UserController{
		Controller: Controller{
			Logger: logger,
			//Storage: storage,
		}}
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

}

func (c *UserController) loginHandler(ctx *gin.Context) {

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
