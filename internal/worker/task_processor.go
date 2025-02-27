package worker

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/almdsh/MyFirstGoApp/internal/models"
)

type TaskProcessor struct {
	db *sql.DB
}

func (tp *TaskProcessor) ProcessTask(task *models.Task) {
	// Update status to in_process
	tp.updateTaskStatus(task.ID, models.StatusInProcess)

	client := &http.Client{
		Timeout: time.Second * 30,
	}

	req, err := http.NewRequest(task.Method, task.URL, nil)
	if err != nil {
		tp.handleTaskError(task.ID, err)
		return
	}

	for k, v := range task.Headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		tp.handleTaskError(task.ID, err)
		return
	}
	defer resp.Body.Close()

	// Update task with response details
	tp.updateTaskComplete(task.ID, resp)
}
