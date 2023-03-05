package controller

import (
	"gorm.io/gorm"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ksusonic/gophermart/internal/api"
	"github.com/ksusonic/gophermart/internal/database"
	"github.com/ksusonic/gophermart/internal/models"
	"github.com/ksusonic/gophermart/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserController struct {
	Controller

	host string
	auth AuthController
}

type AuthController interface {
	IsAuthorized() gin.HandlerFunc
	CreateSignedJWT(claims models.Claims, expiresAt time.Time) (string, error)
}

func NewUserController(host string, auth AuthController, db *database.DB, logger *zap.SugaredLogger) *UserController {
	return &UserController{
		Controller: Controller{
			DB:     db,
			Logger: logger,
		},
		host: host,
		auth: auth,
	}
}

func (c *UserController) RegisterHandlers(router *gin.RouterGroup) {
	router.POST("/register", c.registerHandler)
	router.POST("/login", c.loginHandler)

	authOnly := router.Group("")
	authOnly.Use(c.auth.IsAuthorized())

	authOnly.POST("/orders", c.ordersPostHandler)
	authOnly.GET("/orders", c.ordersGetHandler)
	authOnly.GET("/balance", c.balanceHandler)
	authOnly.POST("/balance/withdraw", c.balanceWithdrawHandler)
	authOnly.GET("/withdrawals", c.withdrawalsHandler)
}

func (c *UserController) registerHandler(ctx *gin.Context) {
	var request api.User

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User

	c.DB.Orm.Where("login = ?", request.Login).First(&existingUser)

	if existingUser.ID != 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "user already exists"})
		return
	}

	hashedPassword, err := utils.GenerateHashPassword(request.Password)
	if err != nil {
		c.Logger.Warnf("could not hash password: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "could not generate password hash"})
		return
	}

	c.DB.Orm.Create(&models.User{
		Login:    request.Login,
		Password: hashedPassword,
	})
	ctx.JSON(http.StatusOK, gin.H{"status": "created"})
}

func (c *UserController) loginHandler(ctx *gin.Context) {
	var request api.User

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	c.DB.Orm.Where("login = ?", request.Login).First(&existingUser)
	if existingUser.ID == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
		return
	}

	errHash := utils.CompareHashPassword(request.Password, existingUser.Password)
	if !errHash {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	expiresAt := time.Now().Add(120 * time.Minute)
	signedToken, err := c.auth.CreateSignedJWT(models.Claims{
		UserID: existingUser.ID,
	}, expiresAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	ctx.SetCookie("Authorization", signedToken, int(expiresAt.Unix()), "/", c.host, false, true)
	ctx.JSON(http.StatusOK, gin.H{"success": "logged in"})
}

func (c *UserController) ordersPostHandler(ctx *gin.Context) {
	userID := getUserIDOrPanic(ctx, c.Logger)

	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect body"})
		return
	}
	orderNumber, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		c.Logger.Error("could not bind request:", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "wrong order format"})
		return
	}

	if !utils.LuhnValid(orderNumber) {
		c.Logger.Info("luhn-invalid number:", orderNumber)
		ctx.JSON(
			http.StatusUnprocessableEntity,
			gin.H{"error": "incorrect order number: " + strconv.FormatInt(orderNumber, 10)},
		)
		return
	}

	var existingOrder models.Order
	err = c.DB.Orm.
		Model(&models.Order{}).
		Where("id = ?", strconv.FormatInt(orderNumber, 10)).
		Find(&existingOrder).
		Error
	if renderIfDBError(ctx, err, "order", c.Logger) {
		return
	}

	// create new order
	if existingOrder.ID == "" {
		err = c.DB.Orm.Model(&models.Order{}).Create(&models.Order{
			ID:     strconv.FormatInt(orderNumber, 10),
			UserID: userID,
			Status: models.StatusNew,
		}).Error
		if renderIfDBError(ctx, err, "order", c.Logger) {
			return
		}
		ctx.JSON(http.StatusAccepted, gin.H{"status": "accepted"})
		return
	} else if existingOrder.UserID == userID {
		ctx.JSON(http.StatusOK, gin.H{"status": "already accepted"})
		return
	} else {
		ctx.JSON(http.StatusConflict, gin.H{"status": "already accepted by another user"})
		return
	}

}

func (c *UserController) ordersGetHandler(ctx *gin.Context) {
	userID := getUserIDOrPanic(ctx, c.Logger)

	var orders []models.Order
	err := c.DB.Orm.Model(&models.Order{}).Where("user_id = ?", userID).Find(&orders).Error
	if renderIfDBError(ctx, err, "order", c.Logger) {
		return
	}

	response := make([]api.Order, len(orders))
	for i := range orders {
		response[i] = api.Order{
			Number:     orders[i].ID,
			Status:     orders[i].Status,
			UploadedAt: orders[i].CreatedAt.Format(time.RFC3339),
		}
		if orders[i].Accrual.Valid {
			response[i].Accrual = orders[i].Accrual.Int64
		}
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *UserController) balanceHandler(ctx *gin.Context) {
	userID := getUserIDOrPanic(ctx, c.Logger)

	var user models.User
	err := c.DB.Orm.Model(&models.User{}).Where("id = ?", userID).Take(&user).Error
	if renderIfDBError(ctx, err, "user", c.Logger) {
		return
	}

	ctx.JSON(http.StatusOK, api.BalanceResponse{
		Current:   user.Current,
		Withdrawn: user.Withdrawn,
	})
}

func (c *UserController) balanceWithdrawHandler(ctx *gin.Context) {
	userID := getUserIDOrPanic(ctx, c.Logger)

	var request api.WithdrawRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	err := c.DB.Orm.Model(&models.Order{}).Where("id = ?", request.Order).Find(&order).Error
	if renderIfDBError(ctx, err, "order", c.Logger) {
		return
	}

	if order.ID == "" {
		ctx.JSON(http.StatusUnprocessableEntity, gin.H{"status": "order not found"})
		return
	}

	var allowedWithdraw bool
	if order.AccrualAvailable >= request.Sum {
		// enough
		allowedWithdraw = true
	} else {
		allowedWithdraw = false
	}

	c.Logger.Infof(
		"user %d with order %s - status %s accrual %d, available %d is %v to withdraw %d for ",
		userID,
		request.Order,
		order.Status,
		order.Accrual.Int64,
		order.AccrualAvailable,
		allowedWithdraw,
		request.Sum,
	)

	if allowedWithdraw {
		err := c.DB.Orm.Transaction(func(tx *gorm.DB) error {
			order.AccrualAvailable -= request.Sum
			if err := tx.Save(order).Error; err != nil {
				return err
			}

			if err := tx.Create(&models.Withdraw{
				OrderID: request.Order,
				Sum:     request.Sum,
			}).Error; err != nil {
				return err
			}

			var user models.User
			if err := tx.Model(&models.User{}).
				Where("id = ?", userID).
				Find(&user).Error; err != nil {
				return err
			}
			user.Current -= float64(request.Sum)
			user.Withdrawn += request.Sum
			if err := tx.Save(user).Error; err != nil {
				return err
			}

			c.Logger.Infof("withdraw of %d is success", request.Sum)
			return nil
		})
		if renderIfDBError(ctx, err, "order update", c.Logger) {
			return
		} else {
			c.Logger.Infof("withdraw of %d is success", request.Sum)
			ctx.JSON(http.StatusOK, gin.H{"status": "success"})
			return
		}
	} else {
		ctx.JSON(http.StatusPaymentRequired, gin.H{"status": "not enough"})
		return
	}
}

func (c *UserController) withdrawalsHandler(ctx *gin.Context) {
	userID := getUserIDOrPanic(ctx, c.Logger)

	var withdrawals []models.Withdraw
	err := c.DB.Orm.Model(&models.Withdraw{}).
		Where("user_id = ?", userID).
		Find(&withdrawals).
		Order("created_at DESC").
		Error
	if renderIfDBError(ctx, err, "withdraw", c.Logger) {
		return
	}

	response := make([]api.WithdrawResponse, len(withdrawals))
	for i := range withdrawals {
		response[i] = api.WithdrawResponse{
			Order:       withdrawals[i].OrderID,
			Sum:         withdrawals[i].Sum,
			ProcessedAt: withdrawals[i].CreatedAt.Format(time.RFC3339),
		}
	}
	ctx.JSON(http.StatusOK, response)
}

func getUserIDOrPanic(ctx *gin.Context, logger *zap.SugaredLogger) uint {
	userID, ok := ctx.Get("user_id")
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal service error"})
		logger.Panicf("no user_id provided by auth middleware! %s %s", ctx.Request.Method, ctx.Request.RequestURI)
	}
	return userID.(uint)
}

func renderIfDBError(ctx *gin.Context, err error, modelName string, logger *zap.SugaredLogger) bool {
	if err != nil {
		logger.Errorf("error retrieving %s: %v", modelName, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error of model " + modelName})
		return true
	}
	return false
}
