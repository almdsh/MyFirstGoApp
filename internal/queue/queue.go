package queue

import (
	"MyFirstGoApp/internal/model"
)

type TaskQueue interface {
	Enqueque(task model.Task)
	Dequeque() model.Task
	Start(num int, process func(model.Task))
	IsEmpty() bool
	Size() int
	Close()
}

type TasksQueue struct {
	tasks chan model.Task
}

func NewTasksQueue(size int) TaskQueue {
	return &TasksQueue{
		tasks: make(chan model.Task, size),
	}
}

func (q *TasksQueue) Enqueque(task model.Task) {
	q.tasks <- task
}

func (q *TasksQueue) Dequeque() model.Task {
	return <-q.tasks
}

func (q *TasksQueue) Start(num int, process func(model.Task)) {
	for i := 0; i < num; i++ {
		go func() {
			for task := range q.tasks {
				process(task)
			}
		}()
	}
}

func (q *TasksQueue) IsEmpty() bool {
	return len(q.tasks) == 0
}

func (q *TasksQueue) Size() int {
	return len(q.tasks)
}
func (q *TasksQueue) Close() {
	close(q.tasks)
}
