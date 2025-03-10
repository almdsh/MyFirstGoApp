package server

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"MyFirstGoApp/internal/database"
	"MyFirstGoApp/internal/model"
)

var DB *sql.DB

func ServerRun() {

	config := database.PostgreSQLConfig{
		Host:     "localhost",
		Port:     "5432",
		Username: "postgresql",
		Password: "postgresql",
		Database: "postgresql",
	}

	DB, err := database.ConnectToDB(config)
	if err != nil {
		log.Fatal(err)
	}

	err = database.CreateTable(DB)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/createTask", createTask)
	mux.HandleFunc("/getTask", getTask)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
}

var Tasks []model.Task

func createTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}
	// Чтение тела запроса
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	task := model.Task{}
	err = json.Unmarshal(body, &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	err = database.AddTask(DB, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task created successfully"))
}

func getTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	tasks, err := database.GetAllTasks(DB)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
