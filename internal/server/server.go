package server

import (
	"MyFirstGoApp/internal/client"
	"MyFirstGoApp/internal/database"
	"MyFirstGoApp/internal/model"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "MyFirstGoApp/docs"

	httpSwagger "github.com/swaggo/http-swagger"
)

type AppContext struct {
	DB *sql.DB
}

func ServerRun() {

	ctx := AppContext{
		DB: database.Run(),
	}
	logSettings()

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tasks", ctx.tasksHandler)
	mux.HandleFunc("/api/v1/tasks/{id}", ctx.taskHandler)
	mux.HandleFunc("/api/v1/swagger/", httpSwagger.WrapHandler)
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", mux))
}

func logSettings() {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open log-file: ", err)
	}
	log.SetOutput(file)
}

func (ctx *AppContext) tasksHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling tasks")
	db := ctx.DB
	switch r.Method {
	case http.MethodPost:
		createTask(w, r, db)

	case http.MethodGet:
		getTasks(w, db)

	case http.MethodDelete:
		deleteTasks(w, db)

	default:
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		log.Println("Invalid request method")
	}
}

func (ctx *AppContext) taskHandler(w http.ResponseWriter, r *http.Request) {
	db := ctx.DB
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "ID is not integer!", http.StatusBadRequest)
		log.Printf("Error parsing ID: %v\n", err)
		return
	}
	log.Printf("Handling task with ID %d\n", id)
	switch r.Method {
	case http.MethodGet:
		getTaskById(w, id, db)
	case http.MethodDelete:
		deleteTaskById(w, id, db)
	default:
		http.Error(w, "Invalid request method", http.StatusBadRequest)
		log.Println("Invalid request method")
	}
}

// @Tags Tasks
// @Router /api/v1/task [post]
// @OperationId createTask
// @Param task body model.Task true "Task object"
// @Summary Create a new task and send it to a third party service
// @Description Creates a new HTTP task
// @Accept json
// @Produce json
// @Success 201 {object} map[string]int64
// @Failure 400 {string} string "Bad request"
// @Failure 500 {string} string "Internal server error"
func createTask(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	task := model.Task{}
	database.UpdateTaskStatus(db, &task, model.New)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error reading request body: %v\n", err)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Printf("Error unmarshaling JSON: %v\n", err)
		return
	}

	var id int64
	id, err = database.AddTask(db, task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error adding task to database: %v\n", err)
		return
	}

	task.ID = id

	log.Printf("Task created successfully with ID %d\n", id)

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int64{"id": id})

	result, err := client.SendTask(db, &task)
	if err != nil {
		log.Printf("Error sending task to third-party service: %v\n", err)
	} else {
		log.Printf("Task with ID %d sent to third-party service successfully\n", id)
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "    ")
	encoder.Encode(result)

}

// @Router /api/v1/tasks [get]
// @OperationId getTasks
// @Summary Get all tasks
// @Description Returns list of all tasks
// @Produce json
// @Success 200 {array} model.Task
// @Failure 500 {string} string "Internal server error"
func getTasks(w http.ResponseWriter, db *sql.DB) {
	tasks, err := database.GetAllTasks(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error getting tasks from database: %v\n", err)
		return
	}

	log.Printf("Tasks retrieved successfully. Count: %d\n", len(tasks))

	jsonData, err := json.Marshal(tasks)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// @Router /api/v1/tasks [delete]
// @OperationId deleteTasks
// @Summary Delete all tasks
// @Description Deletes all tasks from database
// @Success 204 "No Content"
// @Failure 500 {string} string "Internal server error"
func deleteTasks(w http.ResponseWriter, db *sql.DB) {
	err := database.CleanDB(db)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error cleaning database: %v\n", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
	log.Println("Database cleaned successfully")
}

// @Router /api/v1/tasks/{id} [get]
// @OperationId getTaskById
// @Param id path int true "Task ID"
// @Summary Get task by ID
// @Description Returns single task by ID
// @Produce json
// @Success 200 {object} model.Task
// @Failure 404 {string} string "Task not found"
// @Failure 500 {string} string "Internal server error"
func getTaskById(w http.ResponseWriter, id int64, db *sql.DB) {
	task, err := database.GetTaskById(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			log.Printf("Task with ID %d not found\n", id)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Error getting task with ID %d: %v\n", id, err)
		}
		return
	}

	log.Printf("Task with ID %d retrieved successfully\n", id)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(task)
}

// @Router /api/v1/tasks/{id} [delete]
// @OperationId deleteTaskById
// @Summary Delete task
// @Description Deletes task by ID
// @Param id path int true "Task ID"
// @Success 204 "No Content"
// @Failure 404 {string} string "Task not found"
// @Failure 500 {string} string "Internal server error"
func deleteTaskById(w http.ResponseWriter, id int64, db *sql.DB) {
	res, err := database.DeleteTaskById(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Task not found", http.StatusNotFound)
			log.Printf("Task with ID %d not found\n", id)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			log.Printf("Error deleting task with ID %d: %v\n", id, err)
		}
		return
	}
	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("Error getting rows affected: %v\n", err)
		return
	}

	if rows == 0 {
		http.Error(w, "Task not found", http.StatusNotFound)
		log.Printf("Task with ID %d not found\n", id)
		return
	}

	log.Printf("Task with ID %d deleted successfully\n", id)
	w.WriteHeader(http.StatusNoContent)
}
