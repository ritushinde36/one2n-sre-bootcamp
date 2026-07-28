//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/ritushinde36/one2n-sre-bootcamp/connections"
	"github.com/ritushinde36/one2n-sre-bootcamp/controllers"
	"github.com/ritushinde36/one2n-sre-bootcamp/models"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/students", controllers.GetAllStudents)
	router.GET("/api/v1/student/:id", controllers.GetStudent)
	router.POST("/api/v1/students", controllers.CreateStudent)
	router.PUT("/api/v1/student/:id", controllers.UpdateStudent)
	router.DELETE("/api/v1/student/:id", controllers.DeleteStudent)
	return router
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

// TestStudentLifecycle_PersistsAcrossRealDB spins up a real MySQL container
// and drives a student through create -> read -> update -> read -> delete ->
// read, proving each write actually persists in the database rather than
// just returning a success response.
func TestStudentLifecycle_PersistsAcrossRealDB(t *testing.T) {
	ctx := context.Background()

	mysqlContainer, err := tcmysql.Run(ctx,
		"mysql:8.0",
		tcmysql.WithDatabase("student_db"),
		tcmysql.WithUsername("root"),
		tcmysql.WithPassword("rootpass"),
	)
	require.NoError(t, err, "failed to start mysql container")
	t.Cleanup(func() {
		require.NoError(t, testcontainers.TerminateContainer(mysqlContainer))
	})

	dsn, err := mysqlContainer.ConnectionString(ctx, "parseTime=true")
	require.NoError(t, err, "failed to build connection string")

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: connections.NewSlogGormLogger(),
	})
	require.NoError(t, err, "failed to connect to containerized mysql")
	require.NoError(t, db.AutoMigrate(&models.Student{}), "failed to migrate student table")

	connections.DB = db
	router := setupRouter()

	// Create
	createBody := `{"name":"Alice","email":"alice@example.com","age":14,"class":"9th","department":"Music"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusCreated, w.Code, "create should succeed: %s", w.Body.String())

	var created models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotZero(t, created.ID)

	studentPath := "/api/v1/student/" + strconv.FormatUint(uint64(created.ID), 10)

	// Read: the created student should be fetchable with the fields we sent
	w = doRequest(router, http.MethodGet, studentPath, "")
	require.Equal(t, http.StatusOK, w.Code, "get after create should succeed: %s", w.Body.String())

	var fetched models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &fetched))
	assert.Equal(t, "Alice", fetched.Name)
	assert.Equal(t, "alice@example.com", fetched.Email)
	assert.Equal(t, 14, fetched.Age)

	// Update
	updateBody := `{"name":"Alice Updated","email":"alice@example.com","age":15,"class":"10th","department":"Music"}`
	w = doRequest(router, http.MethodPut, studentPath, updateBody)
	require.Equal(t, http.StatusOK, w.Code, "update should succeed: %s", w.Body.String())

	// Read again: the update must have actually persisted, not just returned 200
	w = doRequest(router, http.MethodGet, studentPath, "")
	require.Equal(t, http.StatusOK, w.Code, "get after update should succeed: %s", w.Body.String())

	var updated models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "Alice Updated", updated.Name)
	assert.Equal(t, 15, updated.Age)
	assert.Equal(t, "10th", updated.Class)

	// Delete
	w = doRequest(router, http.MethodDelete, studentPath, "")
	require.Equal(t, http.StatusOK, w.Code, "delete should succeed: %s", w.Body.String())

	// Read again: the delete must have actually persisted, not just returned 200
	w = doRequest(router, http.MethodGet, studentPath, "")
	assert.Equal(t, http.StatusNotFound, w.Code, "student should no longer exist after delete")
}
