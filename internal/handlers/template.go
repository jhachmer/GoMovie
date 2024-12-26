package handlers

import (
	"html/template"
	"log"
	"net/http"
)

var templates *template.Template

func InitTemplates() {
	templates = template.Must(template.ParseFiles("./templates/index.gohtml", "./templates/info.gohtml", "./templates/overview.gohtml", "./templates/movie-grid.gohtml"))
}

func renderTemplate(w http.ResponseWriter, tmpl string, d any) {
	err := templates.ExecuteTemplate(w, tmpl+".gohtml", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
	}
}
