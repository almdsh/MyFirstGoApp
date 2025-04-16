# Task manager server API
## Description
This is a REST API application for managing HTTP tasks.  
The application allows you to create, receive, delete tasks and send HTTP requests to external services.  
The project is covered by unit tests.

### Technologies used
+ Go 1.22.2
+ PostgreSQL
+ Swagger for API documentation
+ Docker Compose

 ## Launching the app
 ### Running in Docker containers
```
 docker-compose up -d
 ```
### Local launch
1. **Install and run PostgreSQL**
   ```
   docker run --rm -d -p 5432:5432 -e POSTGRES_PASSWORD=postgresql -e POSTGRES_USER=postgresql -e POSTGRES_DB=postgresql postgres:latest
    ```
2. **Configure the environment variables:**
```
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgresql
export DB_PASSWORD=postgresql
export DB_NAME=postgresql
```
3. **Launch the application:**
```shell
go run cmd/main.go
```
## API documentation
Swagger UI is available at:
``shell
http://localhost:8080/api/v1/swagger/index.html
``
## Usage examples
1. **Creating a task**
```shell
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"method":"GET","url":"https://google.com","headers":{"Accept":"application/json"}}'
```
2. **Getting all the issues**
```shell
curl -X GET http://localhost:8080/api/v1/tasks
```
3. **Getting an issue by ID**
```shell
curl -X GET http://localhost:8080/api/v1/tasks/1
``
4. **Deleting a task by ID**
```shell
curl -X DELETE http://localhost:8080/api/v1/tasks/1
``
## Project structure
+ **cmd/** - application entry point  
+ **internal/** - internal packages  
  + **server/** - HTTP server and handlers  
  + **database/** - working with the database  
  + **model/** - data models  
  + **client/** - HTTP client for external requests  
+ **docs/** - Swagger documentation
