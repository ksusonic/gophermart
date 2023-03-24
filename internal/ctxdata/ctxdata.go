package ctxdata

import "github.com/gin-gonic/gin"

type ctxKey string

const (
	ctxKeyUserID ctxKey = "user_id"
)

func GetUserID(ctx *gin.Context) (uint, bool) {
	userID, exists := ctx.Get(string(ctxKeyUserID))
	if exists {
		return userID.(uint), exists
	}
	return 0, false
}

func SetUserID(ctx *gin.Context, userID uint) {
	ctx.Set(string(ctxKeyUserID), userID)
}
