package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/ksusonic/gophermart/internal/database"
	"go.uber.org/zap"
)

type Controller struct {
	DB     *database.DB
	Logger *zap.SugaredLogger
}

func (c Controller) RegisterHandlers(*gin.RouterGroup) {
	panic("not implemented!")
}
