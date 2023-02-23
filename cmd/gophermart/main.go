package main

import (
	"log"
	"net/http"

	"github.com/ksusonic/gophermart/internal/config"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	logger := initLogger(cfg.Debug)
	defer logger.Sync()

	if cfg.Debug {
		logger.Debugf("loaded cfg: %s", cfg)
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
	r.Run()
}

func initLogger(debug bool) *zap.SugaredLogger {
	if debug {
		logger, _ := zap.NewDevelopment()
		return logger.Sugar()
	} else {
		logger, _ := zap.NewProduction()
		return logger.Sugar()
	}
}
