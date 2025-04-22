package postgres

import (
	"MyFirstGoApp/internal/model"
	"MyFirstGoApp/internal/storage"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type PostgreSQLConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

type PostgreSQLStorage struct {
	db *sql.DB
}

func GetEnv(env string, fallback string) string {
	value := os.Getenv(env)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func NewPostgreSQLStorage() (storage.Storage, error) {
	config := PostgreSQLConfig{
		Host:     GetEnv("DB_HOST", "db"),
		Port:     GetEnv("DB_PORT", "5432"),
		Username: GetEnv("DB_USER", "postgresql"),
		Password: GetEnv("DB_PASSWORD", "postgresql"),
		Database: GetEnv("DB_NAME", "postgresql"),
	}

	db, err := ConnectToDB(config)
	if err != nil {
		return nil, err //log.Fatal(err)
	}

	err = CreateTable(db)
	if err != nil {
		return nil, err //log.Fatal(err)
	}

	log.Println("Connection to PostgreSQL database established successfully!")
	return &PostgreSQLStorage{db: db}, nil
}

func ConnectToDB(config PostgreSQLConfig) (*sql.DB, error) {
	psqlConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	var err error
	db, err := sql.Open("postgres", psqlConfig)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(`
        DROP TABLE IF EXISTS tasks CASCADE;
    `)
	if err != nil {
		return fmt.Errorf("failed to drop table 'tasks': %w", err)
	}
	_, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS tasks (
            id SERIAL PRIMARY KEY,
            method VARCHAR(10),
            url VARCHAR(255),
            headers JSONB,
            status VARCHAR(20),
            response JSONB
        );
    `)
	if err != nil {
		return fmt.Errorf("failed to create table 'tasks': %w", err)
	}
	return err
}

func (s *PostgreSQLStorage) AddTask(task model.Task) (id int64, err error) {
	headersJSON, err := json.Marshal(task.Headers)
	if err != nil {
		return 0, err
	}

	row := s.db.QueryRow(`
    INSERT INTO tasks (method, url, headers, status)
    VALUES ($1, $2, $3, $4)
    RETURNING id;
    `, task.Method, task.URL, string(headersJSON), task.Status)

	err = row.Scan(&id)
	return id, err
}

func (s *PostgreSQLStorage) GetAllTasks() ([]model.Task, error) {
	rows, err := s.db.Query("SELECT method, url, headers, id, status, response FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		var headers sql.NullString
		var responseJSON sql.NullString
		err = rows.Scan(&task.Method, &task.URL, &headers, &task.ID, &task.Status, &responseJSON)
		if err != nil {
			return nil, err
		}

		if headers.Valid {
			err = json.Unmarshal([]byte(headers.String), &task.Headers)
			if err != nil {
				return nil, err
			}
		}

		if responseJSON.Valid {
			var responseData model.ResponseData
			json.Unmarshal([]byte(responseJSON.String), &responseData)
			task.Response = responseData
		}

		tasks = append(tasks, task)
	}

	return tasks, nil

}

func (s *PostgreSQLStorage) CleanStorage() error {
	_, err := s.db.Exec(`
		TRUNCATE tasks CASCADE;
	`)
	return err
}

func (s *PostgreSQLStorage) GetTaskByID(id int64) (task model.Task, err error) {
	row := s.db.QueryRow("SELECT id, method, url, headers, status, response FROM tasks WHERE id = $1", id)
	var headersJSON, responseJSON sql.NullString
	err = row.Scan(&task.ID, &task.Method, &task.URL, &headersJSON, &task.Status, &responseJSON)
	if err != nil {
		return
	}

	if headersJSON.Valid {
		err = json.Unmarshal([]byte(headersJSON.String), &task.Headers)
		if err != nil {
			return
		}
	}

	if responseJSON.Valid {
		var responseData model.ResponseData
		err = json.Unmarshal([]byte(responseJSON.String), &responseData)
		if err != nil {
			return
		}
		task.Response = responseData
	}

	return task, nil
}

func (s *PostgreSQLStorage) DeleteTaskByID(id int64) (status int64, err error) {
	res, _ := s.db.Exec("DELETE FROM tasks WHERE id = $1", id)

	rows, err := res.RowsAffected()
	if err != nil {
		if err == sql.ErrNoRows {
			status = http.StatusNotFound
		} else {
			status = http.StatusInternalServerError
		}
		return status, err
	}

	if rows == 0 {
		status = http.StatusNotFound
		return status, err
	}

	return status, err
}

func (s *PostgreSQLStorage) UpdateTaskStatus(task *model.Task, status string) error {
	task.Status = status
	_, err := s.db.Exec("UPDATE tasks SET status = $1 WHERE id = $2", status, task.ID)
	if err != nil {
		return fmt.Errorf("error updating task status: %w", err)
	} else {
		log.Printf("Task ID %d status updated to %s\n", task.ID, status)
	}
	return err
}

func (s *PostgreSQLStorage) UpdateTaskResponse(task *model.Task, responseData *model.ResponseData) error {
	responseJSON, err := json.Marshal(responseData)
	if err != nil {
		return err
	}
	_, err = s.db.Exec("UPDATE tasks SET response = $1 WHERE id = $2", string(responseJSON), task.ID)
	return err
}
