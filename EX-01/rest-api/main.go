package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/ritushinde36/one2n-sre-bootcamp/config"
	"github.com/ritushinde36/one2n-sre-bootcamp/connections"
	"github.com/ritushinde36/one2n-sre-bootcamp/controllers"
	"github.com/ritushinde36/one2n-sre-bootcamp/middleware"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	// Reject unknown fields in JSON request bodies (e.g. a client trying to
	// set gorm.Model's ID/CreatedAt/DeletedAt) instead of silently ignoring
	// them.
	binding.EnableDecoderDisallowUnknownFields = true

	//loading the env veriables
	config.LoadConfig()

	// Gin's own package init() reads GIN_MODE from the OS environment before
	// main() ever runs, so it can't see a value that only gets set once
	// config.LoadConfig() parses .env. Re-apply it here now that .env has
	// actually been loaded.
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	// connect to the mysql
	connections.Connect_to_DB()

	//setting up the router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.SlogLogger())

	// setting up the routes
	v1 := router.Group("/api/v1")
	v1.GET("/students", controllers.GetAllStudents)
	v1.GET("/student/:id", controllers.GetStudent)
	v1.POST("/students", controllers.CreateStudent)
	v1.PUT("/student/:id", controllers.UpdateStudent)
	v1.DELETE("/student/:id", controllers.DeleteStudent)

	router.GET("/healthcheck", controllers.HealthCheck)
	router.GET("/readyz", controllers.ReadyCheck)

	//running the router to listen on localhost
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	slog.Info("starting server", "port", port)
	if err := router.Run(":" + port); err != nil {
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
