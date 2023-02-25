package main

import (
	"github.com/ksusonic/gophermart/internal/config"
	"github.com/ksusonic/gophermart/internal/controller"
	"github.com/ksusonic/gophermart/internal/database"
	"github.com/ksusonic/gophermart/internal/server"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	logger := initLogger(cfg.Debug)
	defer logger.Sync()

	db := database.NewDB(cfg.DatabaseURI, logger.Named("DB"))

	s := server.NewServer(cfg, logger)

	s.MountController("/user", controller.NewUserController(
		db,
		logger.Named("user"),
	))
	s.MountController("/orders", controller.NewOrdersController(
		db,
		logger.Named("orders"),
	))

	logger.Fatal(s.Run(cfg.Address))
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
