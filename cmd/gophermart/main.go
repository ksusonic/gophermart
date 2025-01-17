package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ksusonic/gophermart/internal/accrual"
	"github.com/ksusonic/gophermart/internal/auth"
	"github.com/ksusonic/gophermart/internal/config"
	"github.com/ksusonic/gophermart/internal/controller"
	"github.com/ksusonic/gophermart/internal/database"
	"github.com/ksusonic/gophermart/internal/server"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("unable to init config: %v", err)
	}

	logger := initLogger(cfg.Debug)
	defer logger.Sync()

	db, err := database.NewDB(cfg.DatabaseURI, logger.Named("orm"))
	if err != nil {
		log.Fatalf("unable to init DB: %v", err)
	}

	s := server.NewServer(cfg, logger)
	s.MountController("/user", controller.NewUserController(
		auth.NewAuthController(cfg.JwtKey),
		db,
		logger.Named("user"),
	))

	accrualWorker := accrual.NewWorker(
		cfg.AccrualAddress,
		db,
		logger.Named("accrual"),
	)

	ctx, cancel := context.WithCancel(context.Background())
	srv := s.Run(cfg.Address)
	go accrualWorker.Run(ctx)

	defer cancel()

	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	logger.Debugf("caught %v", <-osSignal)

	toCtx, toCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer toCancel()

	if srvErr := srv.Shutdown(toCtx); srvErr != nil {
		logger.Fatalf("s shutdown error: %v", srvErr)
	}

	logger.Info("server stopped")
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
