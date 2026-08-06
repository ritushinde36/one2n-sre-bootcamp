package controllers_test

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ritushinde36/one2n-sre-bootcamp/connections"
	"github.com/ritushinde36/one2n-sre-bootcamp/models"
)

// TestMain starts a single MySQL container shared by every test in this
// package, instead of each test starting its own. Container startup takes a
// few seconds, so sharing it keeps the suite from paying that cost once per
// test function.
func TestMain(m *testing.M) {
	ctx := context.Background()

	mysqlContainer, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("student_db"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("rootpass"),
	)
	if err != nil {
		log.Fatalf("failed to start mysql container: %v", err)
	}

	dsn, err := mysqlContainer.ConnectionString(ctx, "parseTime=true")
	if err != nil {
		log.Fatalf("failed to build connection string: %v", err)
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger:         connections.NewSlogGormLogger(),
	})
	if err != nil {
		log.Fatalf("failed to connect to containerized mysql: %v", err)
	}
	if err := db.AutoMigrate(&models.Student{}); err != nil {
		log.Fatalf("failed to migrate student table: %v", err)
	}

	connections.DB = db

	code := m.Run()

	if err := testcontainers.TerminateContainer(mysqlContainer); err != nil {
		log.Printf("failed to terminate mysql container: %v", err)
	}

	os.Exit(code)
}
