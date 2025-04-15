package database

import (
	"MyFirstGoApp/internal/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
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

func GetEnv(env string, fallback string) string {
	value := os.Getenv(env)
	if len(value) == 0 {
		return fallback
	}
	return value
}
func Run() *sql.DB {
	config := PostgreSQLConfig{
		Host:     GetEnv("DB_HOST", "db"),
		Port:     GetEnv("DB_PORT", "5432"),
		Username: GetEnv("DB_USER", "postgresql"),
		Password: GetEnv("DB_PASSWORD", "postgresql"),
		Database: GetEnv("DB_NAME", "postgresql"),
	}

	db, err := ConnectToDB(config)
	if err != nil {
		log.Fatal(err)
	}

	err = CreateTable(db)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connection to PostgreSQL database established successfully!")
	return db
}

func ConnectToDB(config PostgreSQLConfig) (*sql.DB, error) {
	psqlConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	var err error
	DB, err := sql.Open("postgres", psqlConfig)
	if err != nil {
		return nil, err
	}

	err = DB.Ping()
	if err != nil {
		return nil, err
	}

	return DB, nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(`
        DROP TABLE IF EXISTS tasks CASCADE;
    `)
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
	return err
}

func AddTask(db *sql.DB, task model.Task) (id int64, err error) {
	headersJSON, err := json.Marshal(task.Headers)
	if err != nil {
		return 0, err
	}

	row := db.QueryRow(`
    INSERT INTO tasks (method, url, headers, status)
    VALUES ($1, $2, $3, $4)
    RETURNING id;
    `, task.Method, task.URL, string(headersJSON), task.Status)

	err = row.Scan(&id)
	return id, err
}

func GetAllTasks(db *sql.DB) ([]model.Task, error) {
	rows, err := db.Query("SELECT method, url, headers, id, status, response FROM tasks")
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
			err = json.Unmarshal([]byte(responseJSON.String), &responseData)
			task.Response = responseData
		}

		tasks = append(tasks, task)
	}

	return tasks, nil

}

func CleanDB(db *sql.DB) error {
	_, err := db.Exec(`
		TRUNCATE tasks CASCADE;
	`)
	return err
}

func GetTaskById(db *sql.DB, id int64) (task model.Task, err error) {
	row := db.QueryRow("SELECT id, method, url, headers, status, response FROM tasks WHERE id = $1", id)
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

func DeleteTaskById(db *sql.DB, id int64) (res sql.Result, err error) {
	res, err = db.Exec("DELETE FROM tasks WHERE id = $1", id)
	return
}

func UpdateTaskStatus(db *sql.DB, task *model.Task, status string) {
	task.Status = status
	_, err := db.Exec("UPDATE tasks SET status = $1 WHERE id = $2", status, task.ID)
	if err != nil {
		log.Printf("Error updating task status: %v\n", err)
	} else {
		log.Printf("Task ID %d status updated to %s\n", task.ID, status)
	}
}
