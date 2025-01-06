package handlers

import (
	"html/template"
	"log"
	"net/http"
)

var templates *template.Template

func InitTemplates() {
	templates = template.Must(template.ParseFiles("./templates/index.html",
		"./templates/info.html",
		"./templates/overview.html",
		"./templates/movie-grid.html",
		"./templates/error.html",
		"./templates/register.html"))
}

func renderTemplate(w http.ResponseWriter, tmpl string, d any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
	}
}
