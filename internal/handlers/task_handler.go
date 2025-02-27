package handlers

import (
	"MyFirstGoApp/internal/models"
	"MyFirstGoApp/internal/service"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task := &models.Task{
		ID:      uuid.New().String(),
		Method:  req.Method,
		URL:     req.URL,
		Headers: req.Headers,
		Status:  models.StatusNew,
	}

	if err := h.taskService.CreateTask(r.Context(), task); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"id": task.ID})
}
func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	var req models.GetTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := h.taskService.GetTask(r.Context(), req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(task)
}
