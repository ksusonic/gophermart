package controller

import (
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

	GetUserByLogin(login string, user *models.User) error
	GetOrderByID(id string, order *models.Order) (int64, error)
	GetUserWithdraws(userID uint, withdrawals *[]models.Order) error
	GetOrdersByUserID(userID uint, orders *[]models.Order) error
	CalculateUserStats(userID uint) (userInfo struct{ Balance, Withdraw int64 }, err error)
}
