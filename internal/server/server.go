package server

import (
	"MyFirstGoApp/internal/core"
	"MyFirstGoApp/internal/model"
	"MyFirstGoApp/internal/postgres"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "MyFirstGoApp/docs"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	core *core.App
}

func NewHandlers(app *core.App) *Handlers {
	return &Handlers{
		core: app,
	}
}

func ServerRun() {
	storage, err := postgres.NewPostgreSQLStorage()
	if err != nil {
		log.Fatal(err)
	}
	app := core.NewApp(storage)
	handlers := NewHandlers(app)

	logSettings()

	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()
	router.GET("/api/v1/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.POST("/api/v1/tasks", handlers.createTask)
	router.GET("/api/v1/tasks", handlers.getTasks)
	router.DELETE("/api/v1/tasks", handlers.deleteTasks)
	router.GET("/api/v1/tasks/:id", handlers.getTaskById)
	router.DELETE("/api/v1/tasks/:id", handlers.deleteTaskById)

	router.Run("0.0.0.0:8080")
}

func logSettings() {
	file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal("Failed to open log-file: ", err)
	}
	log.SetOutput(file)
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
func (h *Handlers) createTask(c *gin.Context) {
	var task model.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, result, err := h.core.CreateTask(task)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"id":       id,
		"response": result,
	})
}

// @Router /api/v1/tasks [get]
// @OperationId getTasks
// @Summary Get all tasks
// @Description Returns list of all tasks
// @Produce json
// @Success 200 {array} model.Task
// @Failure 500 {string} string "Internal server error"
func (h *Handlers) getTasks(c *gin.Context) {
	tasks, err := h.core.GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// @Router /api/v1/tasks [delete]
// @OperationId deleteTasks
// @Summary Delete all tasks
// @Description Deletes all tasks from database
// @Success 204 "No Content"
// @Failure 500 {string} string "Internal server error"
func (h *Handlers) deleteTasks(c *gin.Context) {
	err := h.core.CleanStorage()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
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
func (h *Handlers) getTaskById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is not integer!"})
		return
	}

	task, err := h.core.GetTaskByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

// @Router /api/v1/tasks/{id} [delete]
// @OperationId deleteTaskById
// @Summary Delete task
// @Description Deletes task by ID
// @Param id path int true "Task ID"
// @Success 204 "No Content"
// @Failure 404 {string} string "Task not found"
// @Failure 500 {string} string "Internal server error"
func (h *Handlers) deleteTaskById(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is not integer!"})
		return
	}

	status, err := h.core.DeleteTaskByID(id)
	switch status {
	case http.StatusNotFound:
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
	case http.StatusInternalServerError:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	default:
		c.Status(http.StatusNoContent)
	}

	c.Status(http.StatusNoContent)
}
