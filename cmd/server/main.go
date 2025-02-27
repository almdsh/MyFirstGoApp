package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Server struct {
	db     *sql.DB
	router *mux.Router
}

func NewServer(db *sql.DB) *Server {
	s := &Server{
		db:     db,
		router: mux.NewRouter(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.router.HandleFunc("/createTask", s.handlecreateTask).Methods("POST")
	s.router.HandleFunc("/getTask", s.handleGetTask).Methods("GET")
}

func main() {
	// Connect to the database
	connStr := "useer=postgres password=postgres dbname=postgres sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Ping the database to ensure a connection is established
	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	server := NewServer(db)
	log.Printf("server stating on :8080")
	log.Fatal(http.ListenAndServe(":8080", server.router))
}
