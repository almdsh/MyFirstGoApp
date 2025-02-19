package main

import (
	"log"
	"net/http"
)

type LoggingMiddleware struct {
	handler http.Handler
}

func NewLoggingMiddleware(handler http.Handler) *LoggingMiddleware {
	return &LoggingMiddleware{
		handler: handler,
	}
}

func (l *LoggingMiddleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Println(req.URL, "requested")

	l.handler.ServeHTTP(w, req)
}

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_, err := w.Write([]byte("hello"))
	if err != nil {
		log.Println(err)
	}
}

func main() {
	log.Println(http.ListenAndServe(":9090", NewLoggingMiddleware(&Handler{})))
}
