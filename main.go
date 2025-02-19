package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		_, err := w.Write([]byte("hello"))
		if err != nil {
			log.Println(err)
		}
	})

	log.Println("Сервер запущен на http://127.0.0.1:9090/hello")
	log.Println(http.ListenAndServe(":9090", nil))
}
