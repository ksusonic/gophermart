package controller

import (
	"github.com/ksusonic/gophermart/internal/api"
	"github.com/ksusonic/gophermart/internal/models"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Controller struct {
	DB     Database
	Logger *zap.SugaredLogger
}

func (c Controller) RegisterHandlers(*gin.RouterGroup) {
	panic("not implemented!")
}

type Database interface {
	CreateUser(user *models.User) error
	CreateOrder(order *models.Order) error

	GetUserByLogin(login string) (*models.User, error)
	GetOrderByID(id string) (*models.Order, error)
	GetWithdrawnOrdersByUserID(userID uint) (*[]models.Order, error)
	GetOrdersByUserID(userID uint) (*[]models.Order, error)
	CalculateUserStats(userID uint) (*api.UserInfo, error)
}
