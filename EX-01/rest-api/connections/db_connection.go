package connections

import (
	"log/slog"
	"os"

	"github.com/ritushinde36/one2n-sre-bootcamp/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect_to_DB() {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		slog.Error("environment variable DSN is not set")
		os.Exit(1)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         NewSlogGormLogger(),
	})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}

	slog.Info("connected to database")
	DB = db
}

func CreateTable() {
	if err := DB.AutoMigrate(&models.Student{}); err != nil {
		slog.Error("failed to migrate student table", "error", err)
		os.Exit(1)
	}
	slog.Info("student table migrated")
}
