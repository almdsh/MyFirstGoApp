package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/almdsh/MyFirstGoApp/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateTask(t *testing.T) {
	req := models.CreateTaskRequest{
		Method: "GET",
		URL:    "http://example.com",
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	body, _ := json.Marshal(req)
	request := httptest.NewRequest(http.MethodPost, "/createTask", bytes.NewReader(body))
	response := httptest.NewRecorder()

	// Create a test server with a mock DB
	server := createTestServer()
	server.handleCreateTask(response, request)

	assert.Equal(t, http.StatusOK, response.Code)

	var result map[string]string
	json.Unmarshal(response.Body.Bytes(), &result)
	assert.NotEmpty(t, result["id"])
}
