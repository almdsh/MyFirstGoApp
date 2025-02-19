package main

import (
	"html/template"
	"log"
	"net/http"
)

func main() {
	// HTML шаблон с формой
	const tmpl = `
    <!DOCTYPE html>
    <html>
        <body>
            <form method="POST">
                <input type="text" name="message">
                <input type="submit" value="Отправить">
            </form>
            {{if .Message}}
                <h2>Ваше сообщение: {{.Message}}</h2>
            {{end}}
        </body>
    </html>
    `

	t := template.Must(template.New("form").Parse(tmpl))

	http.HandleFunc("/hello", func(w http.ResponseWriter, req *http.Request) {
		data := struct {
			Message string
		}{}

		// Обработка POST запроса
		if req.Method == http.MethodPost {
			data.Message = req.FormValue("message")
		}

		// Отображаем шаблон
		err := t.Execute(w, data)
		if err != nil {
			log.Printf("Template error: %v", err)
		}
	})

	log.Println(http.ListenAndServe(":9090", nil))
}
