package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ritushinde36/one2n-sre-bootcamp/connections"
	"github.com/ritushinde36/one2n-sre-bootcamp/models"
)

// get the records of all the students
func GetAllStudents(c *gin.Context) {
	var all_students []models.Student
	result := connections.DB.Find(&all_students)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to get all student records"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": all_students})

}

// get the record of a particular user
func GetStudent(c *gin.Context) {
	student_id := c.Param("id")
	var student models.Student

	err := connections.DB.First(&student, student_id).Error

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Undable to get the student"})
		return
	}

	c.JSON(http.StatusOK, student)

}

// create a Student
func CreateStudent(c *gin.Context) {
	var new_student models.Student
	err := c.ShouldBindJSON(&new_student)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	connections.DB.Create(&new_student)
	c.JSON(http.StatusCreated, new_student)

}

// Update the record of a student
func UpdateStudent(c *gin.Context) {
	student_id := c.Param("id")
	var student models.Student

	err := connections.DB.First(&student, student_id).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to get the student"})
		return
	}

	var updated_student models.Student

	err = c.ShouldBindJSON(&updated_student)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := connections.DB.Model(&student).Updates(updated_student)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, student)

}

// Delete the record of the student
func DeleteStudent(c *gin.Context) {
	student_id := c.Param("id")
	err := connections.DB.Delete(&models.Student{}, student_id).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "student deleted"})

}
