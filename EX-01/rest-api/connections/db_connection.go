package connections

import (
	"log/slog"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ritushinde36/one2n-sre-bootcamp/config"
)

var DB *gorm.DB

func Connect_to_DB() {
	dsn, err := config.RequireEnv("DSN")
	if err != nil {
		slog.Error(err.Error())
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
