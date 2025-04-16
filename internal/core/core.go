package core

import (
	"MyFirstGoApp/internal/client"
	"MyFirstGoApp/internal/database"
	"MyFirstGoApp/internal/model"
	"database/sql"
	"log"
)

type App struct {
	DB *sql.DB
}

func NewApp(db *sql.DB) *App {
	return &App{
		DB: db,
	}
}

func (a *App) CreateTask(task model.Task) (int64, *model.ResponseData, error) {
	database.UpdateTaskStatus(a.DB, &task, model.New)

	id, err := database.AddTask(a.DB, task)
	if err != nil {
		log.Printf("Error adding task to database: %v\n", err)
		return 0, nil, err
	}

	task.ID = id
	log.Printf("Task created successfully with ID %d\n", id)

	result, err := client.SendTask(a.DB, &task)
	if err != nil {
		log.Printf("Error sending task to third-party service: %v\n", err)
	} else {
		log.Printf("Task with ID %d sent to third-party service successfully\n", id)
	}

	return id, result, err
}

func (a *App) GetAllTasks() ([]model.Task, error) {
	tasks, err := database.GetAllTasks(a.DB)
	if err != nil {
		log.Printf("Error getting tasks from database: %v\n", err)
		return nil, err
	}

	log.Printf("Tasks retrieved successfully. Count: %d\n", len(tasks))
	return tasks, nil
}

func (a *App) DeleteAllTasks() error {
	err := database.CleanDB(a.DB)
	if err != nil {
		log.Printf("Error cleaning database: %v\n", err)
		return err
	}
	log.Println("Database cleaned successfully")
	return nil
}

func (a *App) GetTaskByID(id int64) (model.Task, error) {
	task, err := database.GetTaskById(a.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Task with ID %d not found\n", id)
		} else {
			log.Printf("Error getting task with ID %d: %v\n", id, err)
		}
		return model.Task{}, err
	}

	log.Printf("Task with ID %d retrieved successfully\n", id)
	return task, nil
}

func (a *App) DeleteTaskByID(id int64) (int64, error) {
	res, err := database.DeleteTaskById(a.DB, id)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Task with ID %d not found\n", id)
		} else {
			log.Printf("Error deleting task with ID %d: %v\n", id, err)
		}
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v\n", err)
		return 0, err
	}

	if rows == 0 {
		log.Printf("Task with ID %d not found\n", id)
		return 0, sql.ErrNoRows
	}

	log.Printf("Task with ID %d deleted successfully\n", id)
	return rows, nil
}
