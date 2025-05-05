package HTTPclient

import (
	"MyFirstGoApp/internal/client"
	"MyFirstGoApp/internal/model"
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

func (c *HTTPclient) SendTask(task *model.Task) (*model.ResponseData, error) {

	req, err := http.NewRequest(task.Method, task.URL, bytes.NewBuffer(nil))
	if err != nil {
		log.Println("Request creation error: ", err)
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

	return responseData, err
}
