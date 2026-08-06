package controllers_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ritushinde36/one2n-sre-bootcamp/models"
)

// TestStudentLifecycle_PersistsAcrossRealDB drives a student through
// create -> read -> update -> read -> delete -> read against the shared
// MySQL container (started once in TestMain), proving each write actually
// persists in the database rather than just returning a success response.
func TestStudentLifecycle_PersistsAcrossRealDB(t *testing.T) {
	router := setupRouter()

	// Create
	createBody := `{"name":"Alice","email":"alice-lifecycle@example.com","age":14,"class":"9th","department":"Music"}`
	w := doRequest(router, http.MethodPost, "/api/v1/students", createBody)
	require.Equal(t, http.StatusCreated, w.Code, "create should succeed: %s", w.Body.String())

	var created models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &created))
	require.NotZero(t, created.ID)

	path := studentPath(created.ID)

	// Read: the created student should be fetchable with the fields we sent
	w = doRequest(router, http.MethodGet, path, "")
	require.Equal(t, http.StatusOK, w.Code, "get after create should succeed: %s", w.Body.String())

	var fetched models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &fetched))
	assert.Equal(t, "Alice", fetched.Name)
	assert.Equal(t, "alice-lifecycle@example.com", fetched.Email)
	assert.Equal(t, 14, fetched.Age)

	// Update
	updateBody := `{"name":"Alice Updated","email":"alice-lifecycle@example.com","age":15,"class":"10th","department":"Music"}`
	w = doRequest(router, http.MethodPut, path, updateBody)
	require.Equal(t, http.StatusOK, w.Code, "update should succeed: %s", w.Body.String())

	// Read again: the update must have actually persisted, not just returned 200
	w = doRequest(router, http.MethodGet, path, "")
	require.Equal(t, http.StatusOK, w.Code, "get after update should succeed: %s", w.Body.String())

	var updated models.Student
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &updated))
	assert.Equal(t, "Alice Updated", updated.Name)
	assert.Equal(t, 15, updated.Age)
	assert.Equal(t, "10th", updated.Class)

	// Delete
	w = doRequest(router, http.MethodDelete, path, "")
	require.Equal(t, http.StatusOK, w.Code, "delete should succeed: %s", w.Body.String())

	// Read again: the delete must have actually persisted, not just returned 200
	w = doRequest(router, http.MethodGet, path, "")
	assert.Equal(t, http.StatusNotFound, w.Code, "student should no longer exist after delete")
}
