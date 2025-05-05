package queue

import (
	"MyFirstGoApp/internal/model"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestNewTasksQueue(t *testing.T) {
	q := NewTasksQueue(10)
	if q == nil {
		t.Fatal("Expected non-nil queue")
	}

	if !q.IsEmpty() {
		t.Errorf("New queue should be empty")
	}

	if q.Size() != 0 {
		t.Errorf("Expected size 0, got %d", q.Size())
	}
}

func TestEnqueueDequeue(t *testing.T) {
	q := NewTasksQueue(10)
	task := model.Task{
		ID:      1,
		Method:  "GET",
		URL:     "https://example.com",
		Headers: map[string]string{"Content-Type": "application/json"},
		Status:  model.New,
	}

	q.Enqueque(task)

	if q.IsEmpty() {
		t.Errorf("Queue should not be empty after enqueue")
	}

	if q.Size() != 1 {
		t.Errorf("Expected size 1, got %d", q.Size())
	}

	receivedTask := q.Dequeque()

	if receivedTask.ID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, receivedTask.ID)
	}
	if receivedTask.Method != task.Method {
		t.Errorf("Expected task method %s, got %s", task.Method, receivedTask.Method)
	}
	if receivedTask.URL != task.URL {
		t.Errorf("Expected task URL %s, got %s", task.URL, receivedTask.URL)
	}
	if receivedTask.Status != task.Status {
		t.Errorf("Expected task status %s, got %s", task.Status, receivedTask.Status)
	}
	if !q.IsEmpty() {
		t.Errorf("Queue should be empty after dequeue")
	}
}

func TestStart(t *testing.T) {
	q := NewTasksQueue(10)
	var processedCount int
	var mu sync.Mutex

	q.Start(2, func(task model.Task) {
		mu.Lock()
		processedCount++
		mu.Unlock()
	})

	for i := 1; i <= 5; i++ {
		q.Enqueque(model.Task{
			ID:     int64(i),
			Method: "GET",
			URL:    "https://example.com",
			Status: model.New,
		})
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if processedCount != 5 {
		t.Errorf("Expected 5 processed tasks, got %d", processedCount)
	}
	mu.Unlock()
}

func TestClose(t *testing.T) {
	q := NewTasksQueue(10)
	var processedCount int
	var mu sync.Mutex

	q.Start(2, func(task model.Task) {

		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		processedCount++
		mu.Unlock()
	})

	for i := 1; i <= 5; i++ {
		q.Enqueque(model.Task{
			ID:     int64(i),
			Method: "GET",
			URL:    "https://example.com",
			Status: model.New,
		})
	}

	q.Close()
	time.Sleep(1 * time.Second)

	mu.Lock()
	if processedCount != 5 {
		t.Errorf("Expected 5 processed tasks after close, got %d", processedCount)
	}
	mu.Unlock()

	if !q.IsEmpty() {
		t.Errorf("Queue should be empty after close")
	}
}

func TestConcurrentEnqueueDequeue(t *testing.T) {
	q := NewTasksQueue(100)

	const taskCount = 100

	var wg sync.WaitGroup
	wg.Add(taskCount * 2)

	results := make(chan int64, taskCount)

	for i := 1; i <= taskCount; i++ {
		id := int64(i)
		go func() {
			defer wg.Done()
			q.Enqueque(model.Task{
				ID:     id,
				Method: "GET",
				URL:    "https://example.com",
				Status: model.New,
			})
		}()
	}

	for i := 0; i < taskCount; i++ {
		go func() {
			defer wg.Done()
			task := q.Dequeque()
			results <- task.ID
		}()
	}

	wg.Wait()
	close(results)

	receivedIDs := make(map[int64]bool)
	for id := range results {
		receivedIDs[id] = true
	}

	if len(receivedIDs) != taskCount {
		t.Errorf("Expected %d unique task IDs, got %d", taskCount, len(receivedIDs))
	}
	for i := 1; i <= taskCount; i++ {
		if !receivedIDs[int64(i)] {
			t.Errorf("Task ID %d was not received", i)
		}
	}
}

func TestBlockingEnqueue(t *testing.T) {
	q := NewTasksQueue(2)
	q.Enqueque(model.Task{ID: 1, Method: "GET", URL: "https://example.com"})
	q.Enqueque(model.Task{ID: 2, Method: "GET", URL: "https://example.com"})

	blocked := make(chan bool, 1)

	go func() {
		timer := time.NewTimer(100 * time.Millisecond)
		done := make(chan bool)
		go func() {
			q.Enqueque(model.Task{ID: 3, Method: "GET", URL: "https://example.com"})
			done <- true
		}()
		select {
		case <-done:

			blocked <- false
		case <-timer.C:

			blocked <- true
		}
	}()

	if !<-blocked {
		t.Errorf("Enqueque should block when queue is full")
	}

	q.Dequeque()

	time.Sleep(100 * time.Millisecond)

	if q.Size() != 2 {
		t.Errorf("Expected size 2 after dequeue and blocked enqueue, got %d", q.Size())
	}
}

func TestHTTPTaskProcessing(t *testing.T) {
	q := NewTasksQueue(10)
	var processedCount int
	var mu sync.Mutex

	q.Start(2, func(task model.Task) {
		if task.Method == "" || task.URL == "" {
			t.Errorf("Task should have HTTP method and URL")
		}

		task.Status = model.Done
		task.Response = model.ResponseData{
			Status:        "OK",
			StatusCode:    200,
			Headers:       http.Header{"Content-Type": []string{"application/json"}},
			ContentLength: 42,
			Body:          `{"result": "success"}`,
		}

		mu.Lock()
		processedCount++
		mu.Unlock()
	})

	httpTask := model.Task{
		ID:      1,
		Method:  "POST",
		URL:     "https://api.example.com/data",
		Headers: map[string]string{"Authorization": "Bearer token123"},
		Status:  model.New,
	}
	q.Enqueque(httpTask)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if processedCount != 1 {
		t.Errorf("Expected 1 processed task, got %d", processedCount)
	}
	mu.Unlock()

	q.Close()
}
