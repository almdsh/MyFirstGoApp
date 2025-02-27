package models

type TaskStatus string

const (
	StatusNew       TaskStatus = "new"
	StatusInProcess TaskStatus = "in_process"
	StatusDone      TaskStatus = "done"
	StatusError     TaskStatus = "error"
)

type Task struct {
	ID              string              `json:"id"`
	Method          string              `json:"method"`
	URL             string              `json:"url"`
	Headers         map[string]string   `json:"headers"`
	Status          TaskStatus          `json:"status"`
	HTTPStatusCode  int                 `json:"httpStatusCode,omitempty"`
	ResponseHeaders map[string][]string `json:"headers,omitempty"`
	Length          int64               `json:"length,omitempty"`
}

type CreateTaskRequest struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

type GetTaskRequest struct {
	ID string `json:"id"`
}
