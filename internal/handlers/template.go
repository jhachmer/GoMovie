package handlers

import (
	"html/template"
	"log"
	"net/http"
)

var templates *template.Template

func InitTemplates() {
	funcMap := template.FuncMap{"perc": perc}
	templates = template.Must(template.New("").Funcs(funcMap).ParseFiles(
		"./templates/index.html",
		"./templates/info.html",
		"./templates/overview.html",
		"./templates/movie-grid.html",
		"./templates/error.html",
		"./templates/register.html",
		"./templates/admin.html",
		"./templates/stats.html"))
}

func perc(num1, num2 int) float32 {
	return (float32(num1) / float32(num2)) * 100
}

func renderTemplate(w http.ResponseWriter, tmpl string, d any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
	}
}
