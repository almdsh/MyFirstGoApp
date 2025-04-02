package database

import (
	"MyFirstGoApp/internal/model"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type PostgreSQLConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

func Run() *sql.DB {
	config := PostgreSQLConfig{
		Host:     "db",
		Port:     "5432",
		Username: "postgresql",
		Password: "postgresql",
		Database: "postgresql",
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
		CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			method VARCHAR(10),
			url VARCHAR(255),
			headers JSONB,
			status VARCHAR(20),
			http_status_code INTEGER,
			response_headers JSONB,
			content_length INTEGER
		);
	`)
	return err
}

func AddTask(db *sql.DB, task model.Task) (id int64, err error) {
	headersJSON, err1 := json.Marshal(task.Headers)
	if err1 != nil {
		return 0, err1
	}
	headersJSONStr := string(headersJSON)

	row := db.QueryRow(`
	INSERT INTO tasks (method, url, headers, status)
	VALUES ($1, $2, $3, $4)
	RETURNING id;
`, task.Method, task.URL, headersJSONStr, task.Status)

	err = row.Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func GetAllTasks(db *sql.DB) ([]model.Task, error) {
	rows, err := db.Query("SELECT method, url, headers, id, status FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		var headers sql.NullString
		err = rows.Scan(&task.Method, &task.URL, &headers, &task.ID, &task.Status)
		if err != nil {
			return nil, err
		}

		if headers.Valid {
			err = json.Unmarshal([]byte(headers.String), &task.Headers)
			if err != nil {
				return nil, err
			}
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
	row := db.QueryRow("SELECT id, method, url, headers, status FROM tasks WHERE id = $1", id)
	var headersJSON json.RawMessage
	err = row.Scan(&task.ID, &task.Method, &task.URL, &headersJSON, &task.Status)
	if err != nil {
		return
	}

	var headers map[string]string
	err = json.Unmarshal(headersJSON, &headers)
	if err != nil {
		return
	}
	task.Headers = headers

	return task, err
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
