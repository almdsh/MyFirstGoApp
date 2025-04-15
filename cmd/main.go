package main

import (
	"MyFirstGoApp/internal/server"
)

func main() {
	//@title Task manager Server API
	//@version 1.0
	//@description This is a sample server Task manager server.

	//@host localhost:8080
	//@BasePath /api/v1
	server.ServerRun()
}
