package client

import (
	"MyFirstGoApp/internal/model"
	"MyFirstGoApp/internal/storage"
	"bytes"
	"io"
	"log"
	"net/http"
)

type Client interface {
	SendTask(store storage.Storage, task *model.Task) (*model.ResponseData, error)
}

func NewClient() Client {
	return &HTTPclient{}
}

type HTTPclient struct{}

func (c *HTTPclient) SendTask(storage storage.Storage, task *model.Task) (*model.ResponseData, error) {
	storage.UpdateTaskStatus(task, model.In_process)
	req, err := http.NewRequest(task.Method, task.URL, bytes.NewBuffer(nil))
	if err != nil {
		log.Println("Request creation error: ", err)
		storage.UpdateTaskStatus(task, model.Error)
		return nil, err
	}

	for key, value := range task.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request sending error:", err)
		storage.UpdateTaskStatus(task, model.Error)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("Third-party response for task with ID %d: %v\n", task.ID, resp)
	storage.UpdateTaskStatus(task, model.Done)

	responseData := &model.ResponseData{
		Status:        resp.Status,
		StatusCode:    resp.StatusCode,
		Headers:       resp.Header,
		ContentLength: resp.ContentLength,
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	responseData.Body = string(body)

	storage.UpdateTaskResponse(task, responseData)
	return responseData, err
}
