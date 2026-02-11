package main

import (
	"html/template"
	"log"
	"net/http"
)

type PageData struct {
	Error   string
	Success string
}

func main() {
	tmpl := template.Must(template.ParseFiles("templates/register.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{}

		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")
			confirm := r.FormValue("confirm_password")

			if username == "" || email == "" || password == "" {
				data.Error = "All fields are required"
			} else if password != confirm {
				data.Error = "Passwords do not match"
			} else {
				data.Success = "Account created successfully!"
			}
		}

		tmpl.Execute(w, data)
	})

	log.Println("Server running on http://localhost:8080")

	// 🔴 THIS is the key fix
	log.Fatal(http.ListenAndServe(":8080", nil))
}
