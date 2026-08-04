package controllers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ritushinde36/one2n-sre-bootcamp/models"
)

// TestHealthCheck confirms the liveness endpoint responds without checking
// any external dependency.
func TestHealthCheck(t *testing.T) {
	router := setupRouter()

	w := doRequest(router, http.MethodGet, "/healthcheck", "")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
}

// TestReadyCheck confirms the readiness endpoint reports healthy when the
// database is actually reachable, with a consistent response shape.
func TestReadyCheck(t *testing.T) {
	router := setupRouter()

	w := doRequest(router, http.MethodGet, "/readyz", "")
	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "ok", resp["status"])
	assert.Equal(t, "reachable", resp["database"])
}

// TestGetAllStudents was never invoked by any test before this; confirms
// students that were actually created show up in the list response.
func TestGetAllStudents(t *testing.T) {
	router := setupRouter()

	createBody := `{"name":"Bob","email":"bob-list@example.com","age":20,"class":"11th","department":"CS"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusCreated, w.Code, "create should succeed: %s", w.Body.String())

	var created models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	w = doRequest(router, http.MethodGet, "/api/v1/students", "")
	require.Equal(t, http.StatusOK, w.Code, "list should succeed: %s", w.Body.String())

	var resp struct {
		Message []models.Student `json:"message"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	found := false
	for _, s := range resp.Message {
		if s.ID == created.ID {
			found = true
			assert.Equal(t, "Bob", s.Name)
			assert.Equal(t, "bob-list@example.com", s.Email)
		}
	}
	assert.True(t, found, "created student should appear in GetAllStudents response")
}

// TestGetStudent_NotFound was never invoked by any test before this;
// confirms a nonexistent ID returns 404, not the 400 it used to return.
func TestGetStudent_NotFound(t *testing.T) {
	router := setupRouter()

	w := doRequest(router, http.MethodGet, "/api/v1/student/999999", "")
	assert.Equal(t, http.StatusNotFound, w.Code, "expected 404 for nonexistent student: %s", w.Body.String())
}

// TestDeleteStudent_NotFound was never invoked by any test before this;
// confirms deleting a nonexistent ID returns 404 instead of a false 200.
func TestDeleteStudent_NotFound(t *testing.T) {
	router := setupRouter()

	w := doRequest(router, http.MethodDelete, "/api/v1/student/999999", "")
	assert.Equal(t, http.StatusNotFound, w.Code, "expected 404 for deleting a nonexistent student: %s", w.Body.String())
}

// TestCreateStudent_HappyPath covers the actual "create succeeds and returns
// the created student" case, which the existing validation-only tests never
// exercised.
func TestCreateStudent_HappyPath(t *testing.T) {
	router := setupRouter()

	createBody := `{"name":"Carol","email":"carol-create@example.com","age":19,"class":"12th","department":"Math"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusCreated, w.Code, "create should succeed: %s", w.Body.String())

	var created models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	assert.NotZero(t, created.ID)
	assert.Equal(t, "Carol", created.Name)
	assert.Equal(t, "carol-create@example.com", created.Email)
	assert.Equal(t, 19, created.Age)
}

// TestCreateStudent_DuplicateEmail covers the 409 Conflict branch, which was
// never exercised before this.
func TestCreateStudent_DuplicateEmail(t *testing.T) {
	router := setupRouter()

	createBody := `{"name":"Dave","email":"dave-dup@example.com","age":21,"class":"1st","department":"CS"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusCreated, w.Code, "first create should succeed: %s", w.Body.String())

	w = doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusConflict, w.Code, "duplicate email should be rejected: %s", w.Body.String())

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "a student with this email already exists", resp["error"])
}

// TestUpdateStudent_HappyPath covers the actual "update succeeds and
// persists" case, beyond just rejecting bad input.
func TestUpdateStudent_HappyPath(t *testing.T) {
	router := setupRouter()

	createBody := `{"name":"Eve","email":"eve-update@example.com","age":22,"class":"2nd","department":"Physics"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusCreated, w.Code, "create should succeed: %s", w.Body.String())

	var created models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	path := studentPath(created.ID)

	updateBody := `{"name":"Eve Updated","email":"eve-update@example.com","age":23,"class":"3rd","department":"Physics"}`
	w = doRequest(router, http.MethodPut, path, updateBody)
	require.Equal(t, http.StatusOK, w.Code, "update should succeed: %s", w.Body.String())

	w = doRequest(router, http.MethodGet, path, "")
	require.Equal(t, http.StatusOK, w.Code, "get after update should succeed: %s", w.Body.String())

	var updated models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "Eve Updated", updated.Name)
	assert.Equal(t, 23, updated.Age)
}

// TestUpdateStudent_NotFound covers the 404 branch, which was never
// exercised before this.
func TestUpdateStudent_NotFound(t *testing.T) {
	router := setupRouter()

	updateBody := `{"name":"Ghost","email":"ghost@example.com","age":1,"class":"1st","department":"CS"}`
	w := doRequest(router, http.MethodPut, "/api/v1/student/999999", updateBody)
	assert.Equal(t, http.StatusNotFound, w.Code, "expected 404 for updating a nonexistent student: %s", w.Body.String())
}

// TestCreateStudent_MissingRequiredField covers syntactically valid JSON
// that's simply missing required fields - a different failure path than
// the malformed-JSON cases in StudentController_test.go, which never
// exercised the binding:"required" validation.
func TestCreateStudent_MissingRequiredField(t *testing.T) {
	router := setupRouter()

	w := doRequest(router, http.MethodPost, "/api/v1/students", `{"name": "Alice"}`)

	require.Equal(t, http.StatusBadRequest, w.Code, "expected 400 for missing required fields: %s", w.Body.String())

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "Email")
	assert.Contains(t, resp["error"], "Age")
	assert.Contains(t, resp["error"], "Class")
	assert.Contains(t, resp["error"], "Department")
}

// TestUpdateStudent_MissingRequiredField covers the same gap as
// TestCreateStudent_MissingRequiredField, for the Update path.
func TestUpdateStudent_MissingRequiredField(t *testing.T) {
	router := setupRouter()

	w := doRequest(router, http.MethodPut, "/api/v1/student/999999", `{"name": "Alice"}`)

	require.Equal(t, http.StatusBadRequest, w.Code, "expected 400 for missing required fields: %s", w.Body.String())

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "Email")
	assert.Contains(t, resp["error"], "Age")
	assert.Contains(t, resp["error"], "Class")
	assert.Contains(t, resp["error"], "Department")
}

// TestCreateStudent_RejectsUnknownFields covers the mass-assignment fix - a
// client including a field like ID that doesn't exist on StudentRequest
// should be rejected outright, not silently ignored.
func TestCreateStudent_RejectsUnknownFields(t *testing.T) {
	router := setupRouter()

	body := `{"ID": 111, "name": "sam Smith", "email": "unknown-field-test@example.com", "age": 21, "class": "11th", "department": "Music"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", body)

	require.Equal(t, http.StatusBadRequest, w.Code, "expected 400 for unknown field: %s", w.Body.String())

	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, resp["error"], "ID")
}

// TestUpdateStudent_RejectsUnknownFields covers the same gap for Update.
func TestUpdateStudent_RejectsUnknownFields(t *testing.T) {
	router := setupRouter()

	createBody := `{"name":"Frank","email":"frank-unknown-field@example.com","age":22,"class":"9th","department":"CS"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusCreated, w.Code, "create should succeed: %s", w.Body.String())

	var created models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))

	updateBody := `{"DeletedAt": "2020-01-01T00:00:00Z", "name":"Frank","email":"frank-unknown-field@example.com","age":22,"class":"9th","department":"CS"}`
	w = doRequest(router, http.MethodPut, studentPath(created.ID), updateBody)

	require.Equal(t, http.StatusBadRequest, w.Code, "expected 400 for unknown field: %s", w.Body.String())

	var resp2 map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp2))
	assert.Contains(t, resp2["error"], "DeletedAt")
}
