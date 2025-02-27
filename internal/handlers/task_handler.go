package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/almdsh/MyFirstGoApp/internal/models"
)

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task := &models.Task{
		Method:  req.Method,
		URL:     req.URL,
		Headers: req.Headers,
		Status:  models.StatusNew,
	}

	// Сохраняем задачу в БД и запускаем обработку
	taskID, err := s.saveTask(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Запускаем обработку в фоновом режиме
	go s.taskProcessor.ProcessTask(task)

	json.NewEncoder(w).Encode(map[string]string{"id": taskID})
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request) {
	var req models.GetTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	task, err := s.getTaskFromDB(req.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(task)
}
