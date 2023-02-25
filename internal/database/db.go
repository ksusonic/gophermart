package database

import (
	"go.uber.org/zap"

	"github.com/ksusonic/gophermart/internal/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Orm *gorm.DB
}

func NewDB(dbConnect string, logger *zap.SugaredLogger) *DB {
	db, err := gorm.Open(postgres.Open(dbConnect), &gorm.Config{})
	if err != nil {
		logger.Panic(err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		logger.Fatalf("could not migrate User: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		logger.Fatalf("could not migrate Order: %v", err)
	}
	logger.Debug("successfully migrated")

	return &DB{Orm: db}
}
