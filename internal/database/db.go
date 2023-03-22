package database

import (
	"fmt"
	"github.com/ksusonic/gophermart/internal/models"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Orm *gorm.DB
}

func NewDB(dbConnect string, logger *zap.SugaredLogger) (*DB, error) {
	db, err := gorm.Open(postgres.Open(dbConnect), &gorm.Config{})
	if err != nil {
		logger.Panic(err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("could not migrate User: %v", err)
	}
	if err := db.AutoMigrate(&models.Order{}); err != nil {
		return nil, fmt.Errorf("could not migrate Order: %v", err)
	}
	logger.Debug("successfully migrated")

	return &DB{Orm: db}, nil
}
