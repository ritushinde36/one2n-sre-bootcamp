package connections

import (
	"fmt"
	"log/slog"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ritushinde36/one2n-sre-bootcamp/config"
)

var DB *gorm.DB

func Connect_to_DB() error {
	dsn, err := config.RequireEnv("DSN")
	if err != nil {
		return err
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         NewSlogGormLogger(),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	slog.Info("connected to database")
	DB = db
	return nil
}
