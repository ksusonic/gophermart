package main

import (
	"github.com/ksusonic/gophermart/internal/auth"
	"github.com/ksusonic/gophermart/internal/config"
	"github.com/ksusonic/gophermart/internal/controller"
	"github.com/ksusonic/gophermart/internal/database"
	"github.com/ksusonic/gophermart/internal/server"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		panic(err)
	}

	logger := initLogger(cfg.Debug)
	defer logger.Sync()

	db := database.NewDB(cfg.DatabaseURI, logger.Named("orm"))

	s := server.NewServer(cfg, logger)

	s.MountController("/user", controller.NewUserController(
		cfg.Address,
		auth.NewAuthController(cfg.JwtKey),
		db,
		logger.Named("user"),
	))
	s.MountController("/orders", controller.NewOrdersController(
		db,
		logger.Named("orders"),
	))

	err = s.Run(cfg.Address)
	if err != nil {
		logger.Fatal(err)
	}

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	logger.Debugf("caught %v", <-osSignal)
	logger.Infof("server stopped")
}

func initLogger(debug bool) *zap.SugaredLogger {
	if debug {
		logger, _ := zap.NewDevelopment()
		logger.Level()
		return logger.Sugar()
	} else {
		logger, _ := zap.NewProduction()
		return logger.Sugar()
	}
}
