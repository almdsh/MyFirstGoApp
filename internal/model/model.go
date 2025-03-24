package model

import "net/http"

const (
	Error      = "error"
	In_process = "in_process"
	Done       = "done"
	New        = "new"
)

type Task struct {
	Method   string            `json:"method"`
	URL      string            `json:"url"`
	Headers  map[string]string `json:"headers"`
	ID       int64             `json:"id"`
	Status   string            `json:"status"`
	Response http.Response     `json:"response"`
}
