package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrdersController struct {
	Controller
}

func NewOrdersController(logger *zap.SugaredLogger) *OrdersController {
	return &OrdersController{
		Controller: Controller{
			Logger: logger,
			//Storage: storage,
		}}
}

func (c *OrdersController) RegisterHandlers(router *gin.RouterGroup) {
	router.GET("/:number", c.numberHandler)
}

func (c *OrdersController) numberHandler(ctx *gin.Context) {

}
