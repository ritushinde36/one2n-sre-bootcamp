package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/ritushinde36/one2n-sre-bootcamp/controllers"
)

func setupValidationRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/students", controllers.CreateStudent)
	router.PUT("/student/:id", controllers.UpdateStudent)
	return router
}

func performJSONRequest(router *gin.Engine, method, path, rawBody string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(rawBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestCreateStudent(t *testing.T) {
	tests := []struct {
		name       string
		rawBody    string
		wantStatus int
	}{
		{
			name:       "malformed json - unclosed object",
			rawBody:    `{"name": "Alice"`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed json - trailing comma",
			rawBody:    `{"name": "Alice",}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			rawBody:    ``,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong type for age",
			rawBody:    `{"name": "Alice", "age": "twenty"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not a json object at all",
			rawBody:    `"just a string"`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "array instead of object",
			rawBody:    `[1, 2, 3]`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			router := setupValidationRouter()

			w := performJSONRequest(router, http.MethodPost, "/students", tt.rawBody)

			assert.Equal(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "response should be valid JSON")
			assert.Contains(t, resp, "error", "error response should include an 'error' field")
		})
	}
}

func TestUpdateStudent(t *testing.T) {
	tests := []struct {
		name       string
		rawBody    string
		wantStatus int
	}{
		{
			name:       "malformed json - unclosed object",
			rawBody:    `{"name": "Eve"`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "malformed json - trailing comma",
			rawBody:    `{"name": "Eve",}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty body",
			rawBody:    ``,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong type for age",
			rawBody:    `{"name": "Eve", "age": "twenty-five"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not a json object at all",
			rawBody:    `"just a string"`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "array instead of object",
			rawBody:    `[1, 2, 3]`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			router := setupValidationRouter()

			// The id itself doesn't matter for these cases, because
			// JSON validation now happens before any DB lookup.
			w := performJSONRequest(router, http.MethodPut, "/student/1", tt.rawBody)

			assert.Equal(t, tt.wantStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err, "response should be valid JSON")
			assert.Contains(t, resp, "error", "error response should include an 'error' field")
		})
	}
}
