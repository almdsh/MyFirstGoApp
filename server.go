package main

import (
	"fmt"
	"net/http"
	"github.com/gorilla/mux"
)

type Server struct {
	port   string
}

func NewServer(port string) *Server {
	return &Server{
		port: port,
	}
}

func (s *Server) handleTask (w.http.ResponseWriter, r *http.Request) error {	
	return nil
}

func (s *Server) handleLaunchTask (w.http.ResponseWriter, r *http.Request) error {	
	return nil
}

func (s *Server) handleGetTaskId (w.http.ResponseWriter, r *http.Request) error {	
	return nil
}

func (s *Server) handleRemoveTask (w.http.ResponseWriter, r *http.Request) error {	
	return nil
}
