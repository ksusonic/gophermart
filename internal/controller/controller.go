package controller

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/ksusonic/gophermart/internal/api"
	"github.com/ksusonic/gophermart/internal/models"
	"github.com/ksusonic/gophermart/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const defaultTokenTTL = 120 * time.Minute

type UserController struct {
	Controller

	auth AuthController
}

type AuthController interface {
	AuthMiddleware() gin.HandlerFunc
	CreateSignedJWT(claims models.Claims, expiresAt time.Time) (string, error)
	GetUserID(ctx *gin.Context) (uint, error)
}

func NewUserController(auth AuthController, db Database, logger *zap.SugaredLogger) *UserController {
	return &UserController{
		Controller: Controller{
			DB:     db,
			Logger: logger,
		},
		auth: auth,
	}
}

func (c *UserController) RegisterHandlers(router *gin.RouterGroup) {
	router.POST("/register", c.registerHandler)
	router.POST("/login", c.loginHandler)

	authOnly := router.Group("")
	authOnly.Use(c.auth.AuthMiddleware())

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

	existingUser, err := c.DB.GetUserByLogin(request.Login)
	if renderIfEntityError(ctx, err, c.Logger) {
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
		Login:        request.Login,
		PasswordHash: hashedPassword,
	}
	err = c.DB.CreateUser(&user)
	if renderIfEntityError(ctx, err, c.Logger) {
		return
	}

	expiresAt := time.Now().Add(defaultTokenTTL)
	signedToken, err := c.auth.CreateSignedJWT(models.Claims{
		UserID: user.ID,
	}, expiresAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	ctx.SetCookie("Authorization", signedToken, int(expiresAt.Unix()), "/", "", false, true)
	ctx.JSON(http.StatusOK, gin.H{"status": "welcome"})
}

func (c *UserController) loginHandler(ctx *gin.Context) {
	var request api.User

	if err := ctx.ShouldBindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingUser, err := c.DB.GetUserByLogin(request.Login)
	if renderIfEntityError(ctx, err, c.Logger) {
		return
	}
	if errors.Is(err, sql.ErrNoRows) {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user does not exist"})
		return
	}

	errHash := utils.CompareHashPassword(request.Password, existingUser.PasswordHash)
	if !errHash {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid password"})
		return
	}

	expiresAt := time.Now().Add(defaultTokenTTL)
	signedToken, err := c.auth.CreateSignedJWT(models.Claims{
		UserID: existingUser.ID,
	}, expiresAt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate token"})
		return
	}

	ctx.SetCookie("Authorization", signedToken, int(expiresAt.Unix()), "/", "", false, true)
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

	existingOrder, err := c.DB.GetOrderByID(strconv.FormatInt(orderNumber, 10))
	if renderIfEntityError(ctx, err, c.Logger) {
		return
	}

	if errors.Is(err, sql.ErrNoRows) {
		// create new order
		err = c.DB.CreateOrder(&models.Order{
			ID:     strconv.FormatInt(orderNumber, 10),
			UserID: userID,
			Status: models.OrderStatusNew,
		})
		if renderIfEntityError(ctx, err, c.Logger) {
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

	orders, err := c.DB.GetOrdersByUserID(userID)
	if renderIfEntityError(ctx, err, c.Logger) {
		return
	}

	response := make([]api.Order, len(*orders))
	for i := range *orders {
		response[i] = api.Order{
			Number:     (*orders)[i].ID,
			Status:     (*orders)[i].Status,
			UploadedAt: (*orders)[i].CreatedAt.Format(time.RFC3339),
		}
		if (*orders)[i].Accrual.Valid {
			response[i].Accrual = float64((*orders)[i].Accrual.Int64) / 100
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

	userInfo, err := c.DB.CalculateUserStats(userID)
	if renderIfEntityError(ctx, err, c.Logger) {
		return
	}

	c.Logger.Debugf("currently user %d has %f and withdrawn %f", userID, userInfo.Balance, userInfo.Withdraw)

	ctx.JSON(http.StatusOK, api.BalanceResponse{
		Current:   float64(userInfo.Balance) / 100,
		Withdrawn: float64(userInfo.Withdraw) / 100,
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
			Int64: int64(request.Sum * 100),
			Valid: true,
		},
	}

	_, err = c.DB.GetOrderByID(order.ID)
	if !errors.Is(err, sql.ErrNoRows) {
		ctx.JSON(http.StatusConflict, gin.H{"error": "order already exists"})
		return
	} else if renderIfEntityError(ctx, err, c.Logger) {
		return
	}

	err = c.DB.CreateOrder(&order)
	if renderIfEntityError(ctx, err, c.Logger) {
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

	withdrawals, err := c.DB.GetWithdrawnOrdersByUserID(userID)
	if renderIfEntityError(ctx, err, c.Logger) {
		return
	}

	var response = make(api.WithdrawResponse, len(*withdrawals))
	for i := range *withdrawals {
		response[i] = api.Withdraw{
			Order:       (*withdrawals)[i].ID,
			Sum:         float64((*withdrawals)[i].Withdraw.Int64) / 100,
			ProcessedAt: (*withdrawals)[i].UpdatedAt.Format(time.RFC3339),
		}
	}
	ctx.JSON(http.StatusOK, response)
}

func renderIfEntityError(ctx *gin.Context, err error, logger *zap.SugaredLogger) bool {
	if errors.Is(err, sql.ErrNoRows) {
		return false
	}
	if err != nil {
		logger.Errorf("entity error: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal service error"})
		return true
	}
	return false
}
