package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/ritushinde36/one2n-sre-bootcamp/controllers"
)

func main() {
	//setting up the router
	router := gin.Default()

	// setting up the routes
	router.GET("/students", controllers.GetAllStudents)
	router.GET("/student/:id", controllers.GetStudent)
	router.POST("/students", controllers.CreateStudent)
	router.PUT("/student/:id", controllers.UpdateStudent)
	router.DELETE("/student/:id", controllers.DeleteStudent)

	//running the router to listen on localhost
	router.Run()
	fmt.Println("server is up on localhost")
}
