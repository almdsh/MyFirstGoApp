package client

import (
	"MyFirstGoApp/internal/database"
	"MyFirstGoApp/internal/model"
	"bytes"
	"database/sql"
	"io"
	"log"
	"net/http"
)

func SendTask(db *sql.DB, task *model.Task) (*model.ResponseData, error) {
	database.UpdateTaskStatus(db, task, model.In_process)
	req, err := http.NewRequest(task.Method, task.URL, bytes.NewBuffer(nil))
	if err != nil {
		log.Println("Request creation error: ", err)
		database.UpdateTaskStatus(db, task, model.Error)
		return nil, err
	}

	for key, value := range task.Headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Request sending error:", err)
		database.UpdateTaskStatus(db, task, model.Error)
		return nil, err
	}
	defer resp.Body.Close()
	log.Printf("Third-party response for task with ID %d: %v\n", task.ID, resp)
	database.UpdateTaskStatus(db, task, model.Done)

	responseData := model.ResponseData{
		Status:           resp.Status,
		StatusCode:       resp.StatusCode,
		Proto:            resp.Proto,
		ProtoMajor:       resp.ProtoMajor,
		ProtoMinor:       resp.ProtoMinor,
		Headers:          resp.Header,
		ContentLength:    resp.ContentLength,
		TransferEncoding: resp.TransferEncoding,
		Uncompressed:     resp.Uncompressed,
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	responseData.Body = string(body)

	return &responseData, nil
}
