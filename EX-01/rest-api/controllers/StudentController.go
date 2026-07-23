package controllers

import (
	"log/slog"
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
		slog.Error("failed to fetch students", "error", result.Error)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to get all student records"})
		return
	}
	slog.Info("fetched all students", "count", len(all_students))
	c.JSON(http.StatusOK, gin.H{"message": all_students})

}

// get the record of a particular user
func GetStudent(c *gin.Context) {
	student_id := c.Param("id")
	var student models.Student

	err := connections.DB.First(&student, student_id).Error

	if err != nil {
		slog.Warn("student not found", "student_id", student_id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Undable to get the student"})
		return
	}

	slog.Info("fetched student", "student_id", student_id)
	c.JSON(http.StatusOK, student)

}

// create a Student
func CreateStudent(c *gin.Context) {
	var new_student models.Student
	err := c.ShouldBindJSON(&new_student)
	if err != nil {
		slog.Warn("invalid student payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := connections.DB.Create(&new_student).Error; err != nil {
		slog.Error("failed to create student", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slog.Info("created student", "student_id", new_student.ID)
	c.JSON(http.StatusCreated, new_student)

}

// Update the record of a student
// func UpdateStudent(c *gin.Context) {
// 	student_id := c.Param("id")
// 	var student models.Student

// 	err := connections.DB.First(&student, student_id).Error
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to get the student"})
// 		return
// 	}

// 	var updated_student models.Student

// 	err = c.ShouldBindJSON(&updated_student)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
// 		return
// 	}

// 	result := connections.DB.Model(&student).Updates(updated_student)
// 	if result.Error != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
// 		return
// 	}

// 	c.JSON(http.StatusOK, student)

// }

func UpdateStudent(c *gin.Context) {
	student_id := c.Param("id")

	var updated_student models.Student
	if err := c.ShouldBindJSON(&updated_student); err != nil {
		slog.Warn("invalid student payload", "student_id", student_id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var student models.Student
	err := connections.DB.First(&student, student_id).Error
	if err != nil {
		slog.Warn("student not found", "student_id", student_id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"message": "Unable to get the student"})
		return
	}

	result := connections.DB.Model(&student).Updates(updated_student)
	if result.Error != nil {
		slog.Error("failed to update student", "student_id", student_id, "error", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	slog.Info("updated student", "student_id", student_id)
	c.JSON(http.StatusOK, student)
}

// Delete the record of the student
func DeleteStudent(c *gin.Context) {
	student_id := c.Param("id")
	err := connections.DB.Delete(&models.Student{}, student_id).Error
	if err != nil {
		slog.Error("failed to delete student", "student_id", student_id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	slog.Info("deleted student", "student_id", student_id)
	c.JSON(http.StatusOK, gin.H{"message": "student deleted"})

}

func HealthCheck(c *gin.Context) {
	slog.Debug("healthcheck ping")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
