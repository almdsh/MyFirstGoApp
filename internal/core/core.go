package core

import (
	"MyFirstGoApp/internal/HTTPclient"
	"MyFirstGoApp/internal/model"
	"MyFirstGoApp/internal/queue"
	"MyFirstGoApp/internal/storage"
	"fmt"
	"log"
)

type App struct {
	storage storage.Storage
	q       queue.TaskQueue
}

func NewApp(store storage.Storage) *App {
	q := queue.NewTasksQueue(100)
	return &App{
		storage: store,
		q:       q,
	}
}
func (a *App) Initworkers(num int) {
	a.q.Start(num, func(task model.Task) {
		client := HTTPclient.NewClient()
		err := a.storage.UpdateTaskStatus(&task, model.In_process)
		if err != nil {
			log.Printf("Error updating the status of tasks to in_progress: %v\n", err)
		}
		resp, err := client.SendTask(&task)
		if err != nil {
			log.Printf("Error sending task to third-party service: %v\n", err)
			err := a.storage.UpdateTaskStatus(&task, model.Error)
			if err != nil {
				log.Printf("Error updating the status of tasks to error: %v\n", err)
			}
		} else {
			log.Printf("Task with ID %d sent to third-party service successfully\n", task.ID)
			err := a.storage.UpdateTaskStatus(&task, model.Done)
			if err != nil {
				log.Printf("Error updating the status of tasks to done: %v\n", err)
			}
			err = a.storage.UpdateTaskResponse(&task, resp)
			if err != nil {
				log.Printf("Error updating the response data: %v\n", err)
			}
		}
	})
}

func (a *App) CreateTask(task model.Task) (int64, error) {
	err := a.storage.UpdateTaskStatus(&task, model.New)
	if err != nil {
		return -1, fmt.Errorf("error updating the status of tasks to new: %w", err)
	}

	id, err := a.storage.AddTask(task)
	if err != nil {
		log.Printf("Error adding task to database: %v\n", err)
		return -1, fmt.Errorf("adding task to database error: %w", err)
	}

	task.ID = id
	log.Printf("Task created successfully with ID %d\n", id)

	a.q.Enqueque(task)
	log.Printf("Task with ID %d added to processing queue\n", id)
	return id, err

}

func (a *App) GetAllTasks() ([]model.Task, error) {
	return a.storage.GetAllTasks()
}

func (a *App) CleanStorage() error {
	return a.storage.CleanStorage()
}

func (a *App) GetTaskByID(id int64) (model.Task, error) {
	return a.storage.GetTaskByID(id)
}

func (a *App) DeleteTaskByID(id int64) (int64, error) {
	return a.storage.DeleteTaskByID(id)
}
