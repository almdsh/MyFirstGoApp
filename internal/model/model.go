package model

import "net/http"

const (
	Error      = "error"
	In_process = "in_process"
	Done       = "done"
	New        = "new"
)

type Task struct {
	// @Description HTTP method
	Method string `json:"method"`
	// @Description Target URL
	URL string `json:"url"`
	// @Description HTTP headers
	Headers map[string]string `json:"headers"`
	// @Description Task ID
	ID int64 `json:"id"`
	// @Description Task status
	Status string `json:"status"`
	// @Description HTTP response
	Response http.Response `json:"response"`
}

type ResponseData struct {
	Status           string      `json:"status"`
	StatusCode       int         `json:"status_code"`
	Proto            string      `json:"proto"`
	ProtoMajor       int         `json:"proto_major"`
	ProtoMinor       int         `json:"proto_minor"`
	Headers          http.Header `json:"headers"`
	ContentLength    int64       `json:"content_length"`
	TransferEncoding []string    `json:"transfer_encoding"`
	Body             string      `json:"body"`
	Uncompressed     bool        `json:"uncompressed"`
}
