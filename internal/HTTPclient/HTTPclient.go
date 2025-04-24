package HTTPclient

import (
	"MyFirstGoApp/internal/client"
	"MyFirstGoApp/internal/model"
	"MyFirstGoApp/internal/storage"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func NewClient() client.Client {
	return &HTTPclient{}
}

type HTTPclient struct{}

func (c *HTTPclient) SendTask(storage storage.Storage, task *model.Task) (*model.ResponseData, error) {
	err := storage.UpdateTaskStatus(task, model.In_process)
	if err != nil {
		return nil, fmt.Errorf("error updating the status of tasks to in_progress: %w", err)
	}

	req, err := http.NewRequest(task.Method, task.URL, bytes.NewBuffer(nil))
	if err != nil {
		log.Println("Request creation error: ", err)
		err1 := storage.UpdateTaskStatus(task, model.Error)
		if err1 != nil {
			return nil, fmt.Errorf("error updating the status of tasks to error: %w", err1)
		}
		return nil, fmt.Errorf("request creation error: %w", err)
	}

	for key, value := range task.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request sending error:", err)
		err1 := storage.UpdateTaskStatus(task, model.Error)
		if err1 != nil {
			return nil, fmt.Errorf("error updating the status of tasks to error: %w", err1)
		}
		return nil, fmt.Errorf("request sending error: %w", err)
	}
	defer resp.Body.Close()
	log.Printf("Third-party response for task with ID %d: %v\n", task.ID, resp)

	responseData := &model.ResponseData{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("response body reading error: %w", err)
	}
	responseData.Body = string(body)

	err = storage.UpdateTaskResponse(task, responseData)
	if err != nil {
		log.Printf("Failed to update task response for task with ID %d: %v", task.ID, err)
		err1 := storage.UpdateTaskStatus(task, model.Error)
		if err1 != nil {
			return nil, fmt.Errorf("error updating the status of tasks to error: %w", err1)
		}
		return nil, fmt.Errorf("failed to update task response for task with ID %d: %w", task.ID, err)
	}
	err1 := storage.UpdateTaskStatus(task, model.Done)
	if err1 != nil {
		return nil, fmt.Errorf("error updating the status of tasks to done: %w", err1)
	}
	return responseData, err
}
