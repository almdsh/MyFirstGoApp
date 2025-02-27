package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	resp, err := http.Get("http://localhost:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	fmt.Println("Response status:", resp.Status)
}
