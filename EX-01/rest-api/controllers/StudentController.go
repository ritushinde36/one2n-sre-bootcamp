package controllers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ritushinde36/one2n-sre-bootcamp/connections"
	"github.com/ritushinde36/one2n-sre-bootcamp/models"
	"gorm.io/gorm"
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

	id, err := strconv.ParseUint(student_id, 10, 64)
	if err != nil {
		slog.Warn("invalid student id", "student_id", student_id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	var student models.Student
	if err := connections.DB.First(&student, id).Error; err != nil {
		slog.Warn("student not found", "student_id", student_id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"message": "Unable to get the student"})
		return
	}

	slog.Info("fetched student", "student_id", student_id)
	c.JSON(http.StatusOK, student)

}

// StudentRequest is what a client is allowed to set when creating or
// updating a student - deliberately excludes gorm.Model's ID/CreatedAt/
// UpdatedAt/DeletedAt fields, since models.Student embeds them with no
// json:"-" tags (it's a third-party struct, not ours to add tags to), so
// binding directly into models.Student would let a client set them itself.
type StudentRequest struct {
	Name       string `json:"name" binding:"required"`
	Email      string `json:"email" binding:"required,email"`
	Age        int    `json:"age" binding:"required,gt=0"`
	Class      string `json:"class" binding:"required"`
	Department string `json:"department" binding:"required"`
}

// create a Student
func CreateStudent(c *gin.Context) {
	var req StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("invalid student payload", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	new_student := models.Student{
		Name:       req.Name,
		Email:      req.Email,
		Age:        req.Age,
		Class:      req.Class,
		Department: req.Department,
	}

	if err := connections.DB.Create(&new_student).Error; err != nil {
		slog.Error("failed to create student", "error", err)
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "a student with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to create student"})
		return
	}
	slog.Info("created student", "student_id", new_student.ID)
	c.JSON(http.StatusCreated, new_student)

}

func UpdateStudent(c *gin.Context) {
	student_id := c.Param("id")

	id, err := strconv.ParseUint(student_id, 10, 64)
	if err != nil {
		slog.Warn("invalid student id", "student_id", student_id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	var req StudentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		slog.Warn("invalid student payload", "student_id", student_id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var student models.Student
	err = connections.DB.First(&student, id).Error
	if err != nil {
		slog.Warn("student not found", "student_id", student_id, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"message": "Unable to get the student"})
		return
	}

	updated_student := models.Student{
		Name:       req.Name,
		Email:      req.Email,
		Age:        req.Age,
		Class:      req.Class,
		Department: req.Department,
	}

	result := connections.DB.Model(&student).Updates(updated_student)
	if result.Error != nil {
		slog.Error("failed to update student", "student_id", student_id, "error", result.Error)
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			c.JSON(http.StatusConflict, gin.H{"error": "a student with this email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to update student"})
		return
	}

	slog.Info("updated student", "student_id", student_id)
	c.JSON(http.StatusOK, student)
}

// Delete the record of the student
func DeleteStudent(c *gin.Context) {
	student_id := c.Param("id")

	id, err := strconv.ParseUint(student_id, 10, 64)
	if err != nil {
		slog.Warn("invalid student id", "student_id", student_id, "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid student id"})
		return
	}

	result := connections.DB.Unscoped().Delete(&models.Student{}, id)
	if result.Error != nil {
		slog.Error("failed to delete student", "student_id", student_id, "error", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete student"})
		return
	}
	if result.RowsAffected == 0 {
		slog.Warn("student not found", "student_id", student_id)
		c.JSON(http.StatusNotFound, gin.H{"error": "student not found"})
		return
	}
	slog.Info("deleted student", "student_id", student_id)
	c.JSON(http.StatusOK, gin.H{"message": "student deleted"})

}

// HealthCheck is a liveness check - it only confirms the process itself is
// alive and can respond to HTTP, with no external dependencies checked.
// Safe to use as a Kubernetes liveness probe: a failure here means the
// process itself is broken and should be restarted.
func HealthCheck(c *gin.Context) {
	slog.Debug("healthcheck ping")
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ReadyCheck is a readiness check - it confirms the app can currently serve
// real traffic, including that its database dependency is reachable. Safe
// to use as a Kubernetes readiness probe: a failure here should stop
// traffic being routed here, without restarting the process, since
// restarting won't fix an external dependency being down.
func ReadyCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()

	sqlDB, err := connections.DB.DB()
	if err == nil {
		err = sqlDB.PingContext(ctx)
	}
	if err != nil {
		slog.Error("readycheck failed: database unreachable", "error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "database": "unreachable"})
		return
	}

	slog.Debug("readycheck ping")
	c.JSON(http.StatusOK, gin.H{"status": "ok", "database": "reachable"})
}
