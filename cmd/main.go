package main

import (
	"context"
	"internal/service"
	"log"
	"net/http"
	"os"
	"os/signal"
	"pkg/database"
	"syscall"
	"time"

	"./internal/handlers"
	"github.com/gorilla/mux"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	taskService := service.NewTaskService(db)
	taskHandler := handlers.NewTaskHandler(taskService)

	r := mux.NewRouter()
	r.HandleFunc("/api/v1/tasks", taskHandler.CreateTask).Methods(http.MethodPost)
	r.HandleFunc("/api/v1/tasks/{id}", taskHandler.GetTask).Methods(http.MethodGet)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server shutdown failed: %v", err)
	}
}
