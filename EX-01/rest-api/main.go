package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ritushinde36/one2n-sre-bootcamp/config"
	"github.com/ritushinde36/one2n-sre-bootcamp/connections"
	"github.com/ritushinde36/one2n-sre-bootcamp/controllers"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	//loading the env veriables
	config.LoadConfig()

	// connect to the mysql
	connections.Connect_to_DB()

	//create student table
	connections.CreateTable()

	//setting up the router
	router := gin.Default()

	// setting up the routes
	router.GET("/students", controllers.GetAllStudents)
	router.GET("/student/:id", controllers.GetStudent)
	router.POST("/students", controllers.CreateStudent)
	router.PUT("/student/:id", controllers.UpdateStudent)
	router.DELETE("/student/:id", controllers.DeleteStudent)
	router.GET("/healthcheck", controllers.HealthCheck)

	//running the router to listen on localhost
	slog.Info("starting server", "port", 8888)
	if err := router.Run(":8888"); err != nil {
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
