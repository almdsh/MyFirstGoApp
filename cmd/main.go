package main

import (
	"MyFirstGoApp/internal/server"
)

func main() {
	//@title Task celler Server API
	//@version 1.0
	//@description This is a sample server Task celler server.

	//@contact.name API Support
	//@contact.email alim.dyshekov@axxonsoft.dev

	//@host localhost:8080
	//@BasePath /api/v1
	server.ServerRun()
}
