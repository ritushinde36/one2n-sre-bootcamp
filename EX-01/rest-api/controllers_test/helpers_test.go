package controllers_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/ritushinde36/one2n-sre-bootcamp/controllers"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

// setupRouter builds a router wired to the real controller functions, same
// routes main.go registers, so tests exercise the actual handler code.
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/students", controllers.GetAllStudents)
	router.GET("/api/v1/student/:id", controllers.GetStudent)
	router.POST("/api/v1/students", controllers.CreateStudent)
	router.PUT("/api/v1/student/:id", controllers.UpdateStudent)
	router.DELETE("/api/v1/student/:id", controllers.DeleteStudent)
	router.GET("/healthcheck", controllers.HealthCheck)
	router.GET("/readyz", controllers.ReadyCheck)
	return router
}

func studentPath(id uint) string {
	return "/api/v1/student/" + strconv.FormatUint(uint64(id), 10)
}

func doRequest(router *gin.Engine, method, path, body string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req, _ := http.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}
