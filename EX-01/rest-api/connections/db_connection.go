package connections

import (
	"fmt"
	"log"
	"os"

	"github.com/ritushinde36/one2n-sre-bootcamp/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect_to_DB() {
	dsn := os.Getenv("DSN")
	if dsn == "" {
		log.Fatal("Environment variable 'DSN' is not set")
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to DB : ", err)
	}
	fmt.Print("Successfullly connect to DB")
	DB = db
}

func CreateTable() {
	DB.AutoMigrate(&models.Student{})
}
