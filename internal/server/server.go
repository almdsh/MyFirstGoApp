package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

func ServerRun() {
	mux := http.NewServeMux()
	mux.HandleFunc("/createTask", createTask)
	mux.HandleFunc("/getTask", getTask)
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", mux))
}

type Task struct {
	Method  string            `json:"method"`
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
}

var Tasks []Task
var tasksMutex sync.Mutex

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

	task := Task{}
	err = json.Unmarshal(body, &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	tasksMutex.Lock()
	Tasks = append(Tasks, task)
	tasksMutex.Unlock()

	fmt.Printf("Task created: %+v\n", task)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Task created successfully"))
}

func getTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		return
	}

	tasksMutex.Lock()
	jsonData, err := json.Marshal(Tasks)
	tasksMutex.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)

	fmt.Println("This is the Task.")
}
