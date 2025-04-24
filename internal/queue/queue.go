package queue

import (
	"MyFirstGoApp/internal/model"
	"sync"
)

type TasksQueue struct {
	tasks chan model.Task
	wg    sync.WaitGroup
}

func NewTasksQueue(size int) *TasksQueue {
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
		q.wg.Add(1)
		go func() {
			defer q.wg.Done()
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
	q.wg.Wait()
}
