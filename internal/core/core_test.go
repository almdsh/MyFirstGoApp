package core

import (
	"MyFirstGoApp/internal/model"
	"errors"
	"testing"
)

type MockStorage struct {
	tasks          []model.Task
	addTaskFunc    func(task model.Task) (int64, error)
	updateFunc     func(task *model.Task, status string) error
	updateRespFunc func(task *model.Task, resp *model.ResponseData) error
	getAllFunc     func() ([]model.Task, error)
	getByIDFunc    func(id int64) (model.Task, error)
	deleteFunc     func(id int64) (int64, error)
	cleanFunc      func() error
}

func (m *MockStorage) AddTask(task model.Task) (int64, error) {
	if m.addTaskFunc != nil {
		return m.addTaskFunc(task)
	}
	return 0, nil
}

func (m *MockStorage) UpdateTaskStatus(task *model.Task, status string) error {
	if m.updateFunc != nil {
		return m.updateFunc(task, status)
	}
	return nil
}

func (m *MockStorage) UpdateTaskResponse(task *model.Task, resp *model.ResponseData) error {
	if m.updateRespFunc != nil {
		return m.updateRespFunc(task, resp)
	}
	return nil
}

func (m *MockStorage) GetAllTasks() ([]model.Task, error) {
	if m.getAllFunc != nil {
		return m.getAllFunc()
	}
	return m.tasks, nil
}

func (m *MockStorage) GetTaskByID(id int64) (model.Task, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(id)
	}
	for _, task := range m.tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return model.Task{}, errors.New("task not found")
}

func (m *MockStorage) DeleteTaskByID(id int64) (int64, error) {
	if m.deleteFunc != nil {
		return m.deleteFunc(id)
	}
	return 0, nil
}

func (m *MockStorage) CleanStorage() error {
	if m.cleanFunc != nil {
		return m.cleanFunc()
	}
	m.tasks = []model.Task{}
	return nil
}

type MockTaskQueue struct {
	tasks       []model.Task
	enqueueFunc func(task model.Task)
	dequeueFunc func() model.Task
	startFunc   func(num int, process func(model.Task))
	isEmptyFunc func() bool
	sizeFunc    func() int
	closeFunc   func()
	processFunc func(task model.Task)
}

func (m *MockTaskQueue) Enqueque(task model.Task) {
	if m.enqueueFunc != nil {
		m.enqueueFunc(task)
		return
	}
	m.tasks = append(m.tasks, task)
}

func (m *MockTaskQueue) Dequeque() model.Task {
	if m.dequeueFunc != nil {
		return m.dequeueFunc()
	}
	if len(m.tasks) == 0 {
		return model.Task{}
	}
	task := m.tasks[0]
	m.tasks = m.tasks[1:]
	return task
}

func (m *MockTaskQueue) Start(num int, process func(model.Task)) {
	if m.startFunc != nil {
		m.startFunc(num, process)
		return
	}
	m.processFunc = process
}

func (m *MockTaskQueue) IsEmpty() bool {
	if m.isEmptyFunc != nil {
		return m.isEmptyFunc()
	}
	return len(m.tasks) == 0
}

func (m *MockTaskQueue) Size() int {
	if m.sizeFunc != nil {
		return m.sizeFunc()
	}
	return len(m.tasks)
}

func (m *MockTaskQueue) Close() {
	if m.closeFunc != nil {
		m.closeFunc()
	}
}

func TestNewApp(t *testing.T) {
	mockStorage := &MockStorage{}
	app := NewApp(mockStorage)

	if app == nil {
		t.Fatal("Expected non-nil app")
	}

	if app.storage != mockStorage {
		t.Error("Storage not properly initialized")
	}

	if app.q == nil {
		t.Error("Queue not initialized")
	}
}

func TestInitworkers(t *testing.T) {
	mockStorage := &MockStorage{}
	mockQueue := &MockTaskQueue{}

	app := &App{
		storage: mockStorage,
		q:       mockQueue,
	}

	var startCalled bool
	mockQueue.startFunc = func(num int, process func(model.Task)) {
		startCalled = true
		if num != 3 {
			t.Errorf("Expected 3 workers, got %d", num)
		}
	}

	app.Initworkers(3)

	if !startCalled {
		t.Error("Start method was not called")
	}
}

func TestCreateTask(t *testing.T) {

	t.Run("Success", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}

		mockStorage.updateFunc = func(task *model.Task, status string) error {
			if status != model.New {
				t.Errorf("Expected status %s, got %s", model.New, status)
			}
			return nil
		}

		mockStorage.addTaskFunc = func(task model.Task) (int64, error) {
			return 123, nil
		}

		var enqueueCalled bool
		mockQueue.enqueueFunc = func(task model.Task) {
			enqueueCalled = true
			if task.ID != 123 {
				t.Errorf("Expected task ID 123, got %d", task.ID)
			}
		}

		task := model.Task{
			Method: "GET",
			URL:    "https://example.com",
		}
		id, err := app.CreateTask(task)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if id != 123 {
			t.Errorf("Expected ID 123, got %d", id)
		}
		if !enqueueCalled {
			t.Error("Task was not enqueued")
		}
	})

	t.Run("UpdateStatusError", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		mockStorage.updateFunc = func(task *model.Task, status string) error {
			return errors.New("update error")
		}
		task := model.Task{
			Method: "GET",
			URL:    "https://example.com",
		}
		_, err := app.CreateTask(task)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})

	t.Run("AddTaskError", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		mockStorage.updateFunc = func(task *model.Task, status string) error {
			return nil
		}
		mockStorage.addTaskFunc = func(task model.Task) (int64, error) {
			return 0, errors.New("add task error")
		}
		task := model.Task{
			Method: "GET",
			URL:    "https://example.com",
		}
		_, err := app.CreateTask(task)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestGetAllTasks(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}

		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		expectedTasks := []model.Task{
			{ID: 1, Method: "GET", URL: "https://example.com/1"},
			{ID: 2, Method: "POST", URL: "https://example.com/2"},
		}
		mockStorage.getAllFunc = func() ([]model.Task, error) {
			return expectedTasks, nil
		}
		tasks, err := app.GetAllTasks()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(tasks) != len(expectedTasks) {
			t.Errorf("Expected %d tasks, got %d", len(expectedTasks), len(tasks))
		}
		for i, task := range tasks {
			if task.ID != expectedTasks[i].ID {
				t.Errorf("Expected task ID %d, got %d", expectedTasks[i].ID, task.ID)
			}
		}
	})
	t.Run("Error", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		mockStorage.getAllFunc = func() ([]model.Task, error) {
			return nil, errors.New("get all tasks error")
		}
		_, err := app.GetAllTasks()
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestGetTaskByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		expectedTask := model.Task{ID: 123, Method: "GET", URL: "https://example.com"}
		mockStorage.getByIDFunc = func(id int64) (model.Task, error) {
			if id != 123 {
				t.Errorf("Expected ID 123, got %d", id)
			}
			return expectedTask, nil
		}

		task, err := app.GetTaskByID(123)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if task.ID != expectedTask.ID {
			t.Errorf("Expected task ID %d, got %d", expectedTask.ID, task.ID)
		}
		if task.Method != expectedTask.Method {
			t.Errorf("Expected method %s, got %s", expectedTask.Method, task.Method)
		}
		if task.URL != expectedTask.URL {
			t.Errorf("Expected URL %s, got %s", expectedTask.URL, task.URL)
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		mockStorage.getByIDFunc = func(id int64) (model.Task, error) {
			return model.Task{}, errors.New("task not found")
		}

		_, err := app.GetTaskByID(999)

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestDeleteTaskByID(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		mockStorage.deleteFunc = func(id int64) (int64, error) {
			if id != 123 {
				t.Errorf("Expected ID 123, got %d", id)
			}
			return 1, nil
		}
		count, err := app.DeleteTaskByID(123)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 deleted task, got %d", count)
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		mockStorage.deleteFunc = func(id int64) (int64, error) {
			return 0, errors.New("delete error")
		}
		_, err := app.DeleteTaskByID(999)
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestCleanStorage(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		var cleanCalled bool
		mockStorage.cleanFunc = func() error {
			cleanCalled = true
			return nil
		}

		err := app.CleanStorage()
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if !cleanCalled {
			t.Error("CleanStorage method was not called")
		}
	})
	t.Run("Error", func(t *testing.T) {
		mockStorage := &MockStorage{}
		mockQueue := &MockTaskQueue{}
		app := &App{
			storage: mockStorage,
			q:       mockQueue,
		}
		mockStorage.cleanFunc = func() error {
			return errors.New("clean error")
		}

		err := app.CleanStorage()

		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestInitworkersTaskProcessing(t *testing.T) {
	mockStorage := &MockStorage{}
	mockQueue := &MockTaskQueue{}
	app := &App{
		storage: mockStorage,
		q:       mockQueue,
	}
	testTask := model.Task{
		ID:     123,
		Method: "GET",
		URL:    "https://example.com",
		Status: model.New,
	}
	var processFunc func(model.Task)
	mockQueue.startFunc = func(num int, process func(model.Task)) {
		processFunc = process
	}
	statusUpdates := make(map[string]bool)
	mockStorage.updateFunc = func(task *model.Task, status string) error {
		statusUpdates[status] = true
		return nil
	}
	var responseUpdated bool
	mockStorage.updateRespFunc = func(task *model.Task, resp *model.ResponseData) error {
		responseUpdated = true
		return nil
	}
	app.Initworkers(2)
	if processFunc == nil {
		t.Fatal("Process function was not set")
	}

	processFunc(testTask)

	if !statusUpdates[model.In_process] {
		t.Error("Status was not updated to In_process")
	}
	if !statusUpdates[model.Done] {
		t.Error("Status was not updated to Done")
	}
	if !responseUpdated {
		t.Error("Response was not updated")
	}
}

func TestInitworkersTaskSendError(t *testing.T) {
	mockStorage := &MockStorage{}
	mockQueue := &MockTaskQueue{}
	app := &App{
		storage: mockStorage,
		q:       mockQueue,
	}
	testTask := model.Task{
		ID:     123,
		Method: "GET",
		URL:    "https://invalid-url",
		Status: model.New,
	}
	var processFunc func(model.Task)
	mockQueue.startFunc = func(num int, process func(model.Task)) {
		processFunc = process
	}
	statusUpdates := make(map[string]bool)
	mockStorage.updateFunc = func(task *model.Task, status string) error {
		statusUpdates[status] = true
		return nil
	}
	app.Initworkers(2)
	if processFunc == nil {
		t.Fatal("Process function was not set")
	}

	processFunc(testTask)

	if !statusUpdates[model.In_process] {
		t.Error("Status was not updated to In_process")
	}
	t.Log("Note: Cannot fully test error handling without mocking HTTPclient")
}
