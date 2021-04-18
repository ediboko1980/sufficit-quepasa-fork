package controllers

import (
	"html/template"
	"net/http"

	"github.com/sufficit/sufficit-quepasa-fork/models"
)

type indexData struct {
	PageTitle string
}

// IndexHandler renders route GET "/"
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	_, err := models.GetUser(r)
	if err != nil {
		redirectToLogin(w, r)
		return
	}

	data := indexData{
		PageTitle: "Home",
	}

	templates := template.Must(template.ParseFiles("views/layouts/main.tmpl", "views/index.tmpl"))
	templates.ExecuteTemplate(w, "main", data)
}
