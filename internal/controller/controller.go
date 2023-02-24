package controller

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Storage interface {
}

type Controller struct {
	Logger  *zap.SugaredLogger
	Storage Storage
}

func (c Controller) RegisterHandlers(*gin.RouterGroup) {
	panic("not implemented!")
}
