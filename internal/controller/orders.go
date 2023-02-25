package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ksusonic/gophermart/internal/database"
	"go.uber.org/zap"
)

type OrdersController struct {
	Controller
}

func NewOrdersController(db *database.DB, logger *zap.SugaredLogger) *OrdersController {
	return &OrdersController{
		Controller: Controller{
			DB:     db,
			Logger: logger,
		}}
}

func (c *OrdersController) RegisterHandlers(router *gin.RouterGroup) {
	router.GET("/:number", c.numberHandler)
}

func (c *OrdersController) numberHandler(ctx *gin.Context) {

}
