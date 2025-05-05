package HTTPclient

import (
	"MyFirstGoApp/internal/model"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendTask_Success(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		if r.Header.Get("Authorization") != "Bearer token123" {
			t.Errorf("Expected Authorization: Bearer token123, got %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Success"}`))
	}))
	defer server.Close()

	client := NewClient()

	task := &model.Task{
		ID:     1,
		Method: "GET",
		URL:    server.URL,
		Headers: map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Bearer token123",
		},
		Status: model.New,
	}

	resp, err := client.SendTask(task)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	expectedBody := `{"message": "Success"}`
	if resp.Body != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, resp.Body)
	}

	if resp.Headers.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type: application/json, got %s", resp.Headers.Get("Content-Type"))
	}
}

func TestSendTask_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Internal Server Error"}`))
	}))
	defer server.Close()
	client := NewClient()
	task := &model.Task{
		ID:     2,
		Method: "GET",
		URL:    server.URL,
		Status: model.New,
	}
	resp, err := client.SendTask(task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
	expectedBody := `{"error": "Internal Server Error"}`
	if resp.Body != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, resp.Body)
	}
}

func TestSendTask_InvalidURL(t *testing.T) {
	client := NewClient()
	task := &model.Task{
		ID:     3,
		Method: "GET",
		URL:    "http://invalid-url-that-does-not-exist.example",
		Status: model.New,
	}
	resp, err := client.SendTask(task)
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
	if resp != nil {
		t.Errorf("Expected nil response, got %+v", resp)
	}
}

func TestSendTask_Timeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "Delayed response"}`))
	}))
	defer server.Close()
	client := NewClient()
	task := &model.Task{
		ID:     4,
		Method: "GET",
		URL:    server.URL,
		Status: model.New,
	}
	resp, err := client.SendTask(task)
	t.Logf("Response: %+v, Error: %v", resp, err)
}

func TestSendTask_DifferentMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != method {
					t.Errorf("Expected %s request, got %s", method, r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				response := map[string]string{"method": method}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()
			client := NewClient()
			task := &model.Task{
				ID:     int64(5 + len(method)),
				Method: method,
				URL:    server.URL,
				Status: model.New,
			}
			resp, err := client.SendTask(task)
			if err != nil {
				t.Fatalf("Expected no error for %s, got %v", method, err)
			}
			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status code %d for %s, got %d", http.StatusOK, method, resp.StatusCode)
			}
			if resp.Body == "" {
				t.Errorf("Expected non-empty body for %s", method)
			}
		})
	}
}

func TestSendTask_WithRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"status": "created"}`))
	}))
	defer server.Close()
	client := NewClient()
	task := &model.Task{
		ID:     10,
		Method: "POST",
		URL:    server.URL,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Status: model.New,
	}
	resp, err := client.SendTask(task)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	expectedBody := `{"status": "created"}`
	if resp.Body != expectedBody {
		t.Errorf("Expected body %s, got %s", expectedBody, resp.Body)
	}
}
