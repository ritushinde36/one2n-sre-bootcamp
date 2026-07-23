package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ritushinde36/one2n-sre-bootcamp/config"
	"github.com/ritushinde36/one2n-sre-bootcamp/connections"
	"github.com/ritushinde36/one2n-sre-bootcamp/controllers"
)

func main() {
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

	//running the router to listen on localhost
	router.Run(":8888")
	fmt.Println("server is up on localhost")
}
