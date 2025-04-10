package server

import (
	"MyFirstGoApp/internal/model"
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateTask(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock database: %v", err)
	}
	defer mockDB.Close()

	task := model.Task{
		Method:  "GET",
		URL:     "https://example.com",
		Headers: map[string]string{},
		Status:  model.New,
	}
	task.Headers["Content-Type"] = "application/json"
	headersJSON, _ := json.Marshal(task.Headers)

	mock.ExpectQuery("INSERT INTO tasks \\(method, url, headers, status\\) VALUES \\(\\$1, \\$2, \\$3, \\$4\\) RETURNING id").
		WithArgs(task.Method, task.URL, string(headersJSON), task.Status).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	body, err := json.Marshal(task)
	if err != nil {
		t.Fatalf("Error marshaling task: %v", err)
	}
	request, err := http.NewRequest("POST", "/api/v1/tasks", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Error creating request: %v", err)
	}

	w := httptest.NewRecorder()

	createTask(w, request, mockDB)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, w.Code)
	}

	var responseBytes []byte
	responseBytes, err = io.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	firstResponse := strings.Split(string(responseBytes), "\n")[0]

	var response map[string]int64
	if err := json.Unmarshal([]byte(firstResponse), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %s", err)
	}
}

func TestGetTasks(t *testing.T) {
	testCases := []struct {
		name               string
		mockSetup          func(mock sqlmock.Sqlmock)
		expectedHTTPStatus int
		expectedTask       []model.Task
		expectedStatus     int
		expectedError      bool // error?
	}{
		{
			name: "Successfully retrieve empty tasks list",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"method", "url", "headers", "id", "status"})
				mock.ExpectQuery("SELECT method, url, headers, id, status FROM tasks").
					WillReturnRows(rows)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedTask:       []model.Task{},
			expectedError:      false,
		},
		{
			name: "Successfully retrieve  tasks list with many task",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"method", "url", "headers", "id", "status"}).
					AddRow("GET", "https://example.com", `{"Content-Type":"application/json"}`, 1, "new").
					AddRow("POST", "https://api.example.com", `{"Authorization": "token"}`, 2, "in_progress").
					AddRow("PUT", "https://update.example.com", `{"Custom-Header":"value"}`, 3, "done")
				mock.ExpectQuery("SELECT method, url, headers, id, status FROM tasks").
					WillReturnRows(rows)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedTask: []model.Task{
				{
					Method: "GET",
					URL:    "https://example.com",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					ID:     1,
					Status: "new",
				},
				{
					Method: "POST",
					URL:    "https://api.example.com",
					Headers: map[string]string{
						"Authorization": "token",
					},
					ID:     2,
					Status: "in_progress",
				},
				{
					Method: "PUT",
					URL:    "https://update.example.com",
					Headers: map[string]string{
						"Custom-Header": "value",
					},
					ID:     3,
					Status: "done",
				},
			},
			expectedError: false,
		},
		{
			name: "Database error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT method, url, headers, id, status FROM tasks").
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedError:      true,
		},
		{
			name: "Invalid JSON in headers",
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"method", "url", "headers", "id"}).
					AddRow("GET", "https://example.com", `{invalid-json}`, 1)
				mock.ExpectQuery("SELECT method, url, headers, id, status FROM tasks").
					WillReturnRows(rows)
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedError:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("Error creating mock database: %v", err)
			}
			defer mockDB.Close()

			tc.mockSetup(mock)

			w := httptest.NewRecorder()

			getTasks(w, mockDB)

			if w.Code != tc.expectedHTTPStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedHTTPStatus, w.Code)
			}

			if !tc.expectedError {
				var tasks []model.Task
				err = json.Unmarshal(w.Body.Bytes(), &tasks)
				if err != nil {
					t.Errorf("Error unmarshalling JSON: %v", err)
				}

				if len(tasks) != len(tc.expectedTask) {
					t.Errorf("Expected %d tasks, got %d", len(tc.expectedTask), len(tasks))
				}

				for i, expectedTask := range tc.expectedTask {

					if i >= len(tasks) {
						break
					}
					actualTask := tasks[i]

					assert := assert.New(t)
					assert.Equal(expectedTask, actualTask)

					for key, expectedValue := range expectedTask.Headers {
						actualValue, exists := actualTask.Headers[key]
						if !exists {
							t.Errorf("Task %d: Header %s not found", i, key)
						} else if actualValue != expectedValue {
							t.Errorf("Task %d: Header %s expected value %s, got %s", i, key, expectedValue, actualValue)
						}
					}
				}
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestGetTaskById(t *testing.T) {
	testCases := []struct {
		name               string
		taskID             int64
		mockSetup          func(mock sqlmock.Sqlmock)
		expectedHTTPStatus int
		expectedTask       *model.Task
		expectedError      bool
	}{
		{
			name:   "Successfully retrieve task by ID",
			taskID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				task := model.Task{
					Method:  "GET",
					URL:     "https://example.com",
					Headers: map[string]string{},
					Status:  "new",
				}
				task.Headers["Content-Type"] = "application/json"
				headersJSON, _ := json.Marshal(task.Headers)
				rows := sqlmock.NewRows([]string{"id", "method", "url", "headers", "status"}).
					AddRow(1, task.Method, task.URL, headersJSON, task.Status)

				mock.ExpectQuery("SELECT id, method, url, headers, status FROM tasks WHERE id = \\$1").
					WithArgs(1).
					WillReturnRows(rows)
			},
			expectedHTTPStatus: http.StatusOK,
			expectedTask: &model.Task{
				ID:     1,
				Method: "GET",
				URL:    "https://example.com",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Status: "new",
			},
			expectedError: false,
		},
		{
			name:   "Task not found",
			taskID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, method, url, headers, status FROM tasks WHERE id = \\$1").
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			expectedHTTPStatus: http.StatusNotFound,
			expectedTask:       nil,
			expectedError:      true,
		},
		{
			name:   "Database error",
			taskID: 2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, method, url, headers, status FROM tasks WHERE id = \\$1").
					WithArgs(2).
					WillReturnError(fmt.Errorf("database connection error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedTask:       nil,
			expectedError:      true,
		},
		{
			name:   "Invalid JSON in headers",
			taskID: 3,
			mockSetup: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "method", "url", "headers", "status"}).
					AddRow(3, "GET", "https://example.com", `{invalid-json}`, "new")
				mock.ExpectQuery("SELECT id, method, url, headers, status FROM tasks WHERE id = \\$1").
					WithArgs(3).
					WillReturnRows(rows)
			},
			expectedHTTPStatus: http.StatusInternalServerError,
			expectedTask:       nil,
			expectedError:      true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			require.NoError(t, err, "Error creating mock database")
			defer mockDB.Close()

			tc.mockSetup(mock)

			w := httptest.NewRecorder()

			getTaskById(w, tc.taskID, mockDB)

			assert.Equal(t, tc.expectedHTTPStatus, w.Code)

			if !tc.expectedError && tc.expectedTask != nil {
				var task model.Task
				err = json.NewDecoder(w.Body).Decode(&task) //(w.Body.Bytes(), &task)
				require.NoError(t, err, "Error unmarshalling JSON")

				assert.Equal(t, *tc.expectedTask, task)

				for key, expectedValue := range tc.expectedTask.Headers {
					actualValue, exists := task.Headers[key]
					assert.True(t, exists, "Header %s not found", key)
					if exists {
						assert.Equal(t, expectedValue, actualValue, "Header %s value mismatch", key)
					}
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet(), "Unfulfilled expectations")
		})
	}
}

func TestDeleteTaskById(t *testing.T) {
	testCases := []struct {
		name               string
		taskID             int64
		mockSetup          func(mock sqlmock.Sqlmock)
		expectedHTTPStatus int
	}{
		{
			name:   "Successfully delete task",
			taskID: 1,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
					WithArgs(1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			expectedHTTPStatus: http.StatusNoContent,
		},
		{
			name:   "Task not found",
			taskID: 999,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
					WithArgs(999).
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedHTTPStatus: http.StatusNotFound,
		},
		{
			name:   "Database error",
			taskID: 2,
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM tasks WHERE id = \\$1").
					WithArgs(2).
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer mockDB.Close()

			tc.mockSetup(mock)
			w := httptest.NewRecorder()

			deleteTaskById(w, tc.taskID, mockDB)

			assert.Equal(t, tc.expectedHTTPStatus, w.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteTasks(t *testing.T) {
	testCases := []struct {
		name               string
		mockSetup          func(mock sqlmock.Sqlmock)
		expectedHTTPStatus int
	}{
		{
			name: "Successfully delete all tasks",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("TRUNCATE tasks CASCADE").
					WillReturnResult(sqlmock.NewResult(0, 0))
			},
			expectedHTTPStatus: http.StatusNoContent,
		},
		{
			name: "Database error",
			mockSetup: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("TRUNCATE tasks CASCADE").
					WillReturnError(fmt.Errorf("database error"))
			},
			expectedHTTPStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer mockDB.Close()

			tc.mockSetup(mock)
			w := httptest.NewRecorder()

			deleteTasks(w, mockDB)

			assert.Equal(t, tc.expectedHTTPStatus, w.Code)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestTaskHandler(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		url                string
		expectedHTTPStatus int
	}{
		{
			name:               "Invalid ID format",
			method:             http.MethodGet,
			url:                "/tasks/invalid",
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name:               "Invalid HTTP method",
			method:             http.MethodPatch,
			url:                "/tasks/1",
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, _, err := sqlmock.New()
			require.NoError(t, err)
			defer mockDB.Close()

			ctx := &AppContext{DB: mockDB}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, tc.url, nil)

			ctx.taskHandler(w, r)
			assert.Equal(t, tc.expectedHTTPStatus, w.Code)
		})
	}
}

func TestTasksHandler(t *testing.T) {
	testCases := []struct {
		name               string
		method             string
		expectedHTTPStatus int
	}{
		{
			name:               "Invalid HTTP method",
			method:             http.MethodPatch,
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockDB, _, err := sqlmock.New()
			require.NoError(t, err)
			defer mockDB.Close()

			ctx := &AppContext{DB: mockDB}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tc.method, "/tasks", nil)

			ctx.tasksHandler(w, r)
			assert.Equal(t, tc.expectedHTTPStatus, w.Code)
		})
	}
}
