package client

import (
	"MyFirstGoApp/internal/database"
	"MyFirstGoApp/internal/model"
	"bytes"
	"database/sql"
	"log"
	"net/http"
)

func SendTask(db *sql.DB, task *model.Task) {
	database.UpdateTaskStatus(db, task, model.In_process)
	req, err := http.NewRequest(task.Method, task.URL, bytes.NewBuffer(nil))
	if err != nil {
		log.Println("Request creation error: ", err)
		database.UpdateTaskStatus(db, task, model.Error)
		return
	}

	for key, value := range task.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request sending error:", err)
		database.UpdateTaskStatus(db, task, model.Error)
		return
	}
	defer resp.Body.Close()
	log.Printf("Third-party response for task with ID %d: %v\n", task.ID, resp)
	database.UpdateTaskStatus(db, task, model.Done)
}
