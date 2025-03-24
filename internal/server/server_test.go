package server

import (
	"MyFirstGoApp/internal/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateTask(t *testing.T) {
	//
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer mockDB.Close()

	task := model.Task{
		Method:  "GET",
		URL:     "https://example.com",
		Headers: map[string]string{},
	}

	headersJSON, _ := json.Marshal(task.Headers)
	mock.ExpectQuery("INSERT INTO tasks \\(method, url, headers, status\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) RETURNING id").
		WithArgs(task.Method, task.URL, string(headersJSON), task.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	body, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Error marshaling task: %v", err)
	}
	request, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}
	w := httptest.NewRecorder()

	//
	createTask(w, request, mockDB)

	//
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]int64
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["id"] != 1 {
		t.Errorf("Expected ID 1, got %d", response["id"])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}
