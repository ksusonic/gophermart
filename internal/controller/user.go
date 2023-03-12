package controller

import (
	"database/sql"
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
	GetUserID(ctx *gin.Context) (uint, error)
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

	err := c.DB.Orm.Where("login = ?", request.Login).Limit(1).Find(&existingUser).Error
	if renderIfDBError(ctx, err, "order", c.Logger) {
		return
	}

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

	user := models.User{
		Login:    request.Login,
		Password: hashedPassword,
	}
	err = c.DB.Orm.Create(&user).Error
	if renderIfDBError(ctx, err, "order", c.Logger) {
		return
	}

	expiresAt := time.Now().Add(120 * time.Minute)
	signedToken, err := c.auth.CreateSignedJWT(models.Claims{
		UserID: user.ID,
	}, expiresAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	ctx.SetCookie("Authorization", signedToken, int(expiresAt.Unix()), "/", c.host, false, true)
	ctx.JSON(http.StatusOK, gin.H{"status": "welcome"})
}

func (c *UserController) loginHandler(ctx *gin.Context) {
	var request api.User

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existingUser models.User
	err := c.DB.Orm.Where("login = ?", request.Login).Limit(1).Find(&existingUser).Error
	if renderIfDBError(ctx, err, "order", c.Logger) {
		return
	}
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
	userID, err := c.auth.GetUserID(ctx)
	if err != nil {
		c.Logger.Errorf("not found user_id in context: %s %s", ctx.Request.Method, ctx.Request.RequestURI)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal service error"})
		return
	}

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
			Status: models.OrderStatusNew,
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
	userID, err := c.auth.GetUserID(ctx)
	if err != nil {
		c.Logger.Errorf("not found user_id in context: %s %s", ctx.Request.Method, ctx.Request.RequestURI)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal service error"})
		return
	}

	var orders []models.Order
	err = c.DB.Orm.Model(&models.Order{}).Where("user_id = ?", userID).Find(&orders).Error
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
	userID, err := c.auth.GetUserID(ctx)
	if err != nil {
		c.Logger.Errorf("not found user_id in context: %s %s", ctx.Request.Method, ctx.Request.RequestURI)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal service error"})
		return
	}

	userInfo := struct {
		Balance  int64
		Withdraw int64
	}{}

	err = c.DB.Orm.
		Table("orders").
		Select("sum(accrual)-sum(withdraw) as balance, sum(withdraw) as withdraw").
		Where("user_id=?", userID).
		Scan(&userInfo).
		Error
	if renderIfDBError(ctx, err, "orders", c.Logger) {
		return
	}

	c.Logger.Debugf("currently user %d has %d and withdrawn %d", userID, userInfo.Balance, userInfo.Withdraw)

	ctx.JSON(http.StatusOK, api.BalanceResponse{
		Current:   userInfo.Balance,
		Withdrawn: userInfo.Withdraw,
	})
}

func (c *UserController) balanceWithdrawHandler(ctx *gin.Context) {
	userID, err := c.auth.GetUserID(ctx)
	if err != nil {
		c.Logger.Errorf("not found user_id in context: %s %s", ctx.Request.Method, ctx.Request.RequestURI)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal service error"})
		return
	}

	var request api.WithdrawRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order = models.Order{
		ID:     request.Order,
		UserID: userID,
		Status: models.OrderStatusNew,
		Withdraw: sql.NullInt64{
			Int64: request.Sum,
			Valid: true,
		},
	}

	findTx := c.DB.Orm.Model(&models.Order{}).Where("id = ?", order.ID).Limit(1).Find(&models.Order{})
	if renderIfDBError(ctx, findTx.Error, "orders", c.Logger) {
		return
	} else if findTx.RowsAffected != 0 {
		ctx.JSON(http.StatusConflict, gin.H{"error": "order already exists"})
		return
	}

	err = c.DB.Orm.Create(&order).Error
	if renderIfDBError(ctx, err, "orders", c.Logger) {
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "ok - withdrawn"})
}

func (c *UserController) withdrawalsHandler(ctx *gin.Context) {
	userID, err := c.auth.GetUserID(ctx)
	if err != nil {
		c.Logger.Errorf("not found user_id in context: %s %s", ctx.Request.Method, ctx.Request.RequestURI)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal service error"})
		return
	}

	var withdrawals []models.Order
	err = c.DB.Orm.Model(&models.Order{}).
		Where("user_id = ? and withdraw not null", userID).
		Take(&withdrawals).
		Error
	if renderIfDBError(ctx, err, "orders", c.Logger) {
		return
	}

	var response = make(api.WithdrawResponse, len(withdrawals))
	for i := range withdrawals {
		response[i] = api.Withdraw{
			Order:       withdrawals[i].ID,
			Sum:         withdrawals[i].Withdraw.Int64,
			ProcessedAt: withdrawals[i].UpdatedAt.Format(time.RFC3339),
		}
	}
	ctx.JSON(http.StatusOK, response)
}

func renderIfDBError(ctx *gin.Context, err error, modelName string, logger *zap.SugaredLogger) bool {
	if err != nil {
		logger.Errorf("error retrieving %s: %v", modelName, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error of model " + modelName})
		return true
	}
	return false
}
