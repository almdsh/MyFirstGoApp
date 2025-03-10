package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"

	"MyFirstGoApp/internal/model"

	_ "github.com/lib/pq"
)

type PostgreSQLConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	Database string
}

//var DB *sql.DB

func ConnectToDB(config PostgreSQLConfig) (*sql.DB, error) {
	psqlConfig := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	var err error
	DB, err := sql.Open("postgres", psqlConfig)
	if err != nil {
		return DB, err
	}

	err = DB.Ping()
	if err != nil {
		return DB, err
	}

	log.Println("Connection to PostgreSQL database")
	return DB, nil
}

func CreateTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id SERIAL PRIMARY KEY,
			method VARCHAR(10),
			url VARCHAR(255),
			headers JSONB
		);
	`)
	return err
}

func AddTask(db *sql.DB, task model.Task) error {
	_, err := db.Exec(`
		INSERT INFO tasks (method, url, headers)
		VALUES ($1, $2, $3);
	`, task.Method, task.URL, task.Headers)
	return err
}

func GetAllTasks(db *sql.DB) ([]model.Task, error) {
	rows, err := db.Query("SELECT method, url, headers FROM tasks")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var task model.Task
		var headers sql.NullString
		err = rows.Scan(&task.Method, &task.URL, &headers)
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
