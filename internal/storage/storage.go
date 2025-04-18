package storage

import (
	"MyFirstGoApp/internal/model"
)

type Storage interface {
	AddTask(task model.Task) (int64, error)
	GetAllTasks() ([]model.Task, error)
	GetTaskByID(id int64) (model.Task, error)
	DeleteTaskByID(id int64) (int64, error)
	UpdateTaskStatus(task *model.Task, status string) error
	UpdateTaskResponse(task *model.Task, response *model.ResponseData) error
	CleanStorage() error
}
