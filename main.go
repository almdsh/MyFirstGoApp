package main

import (
	"fmt"
	"log"
	"net/http"
)

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	_, err := w.Write([]byte("Hello, world\n"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(w, "Hello, you've requested: %s\n", req.URL.Path)
}

func main() {
	log.Println(http.ListenAndServe("127.0.0.1:9090", &Handler{}))

}
