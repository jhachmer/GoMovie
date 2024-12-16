package server

import (
	"fmt"
	"github.com/jhachmer/gotocollection/pkg/media"
	"github.com/jhachmer/gotocollection/pkg/store"
	"github.com/jhachmer/gotocollection/pkg/types"
	"github.com/jhachmer/gotocollection/pkg/util"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

type Handler struct {
	store *store.Storage
}

func NewHandler(store *store.Storage) *Handler {
	return &Handler{
		store: store,
	}
}

var templates = template.Must(template.ParseFiles("./templates/index.gohtml", "./templates/info.gohtml"))
var validPath = regexp.MustCompile("^/films/([a-zA-Z0-9]+)$")

func (h Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := util.Encode(w, r, http.StatusOK, map[string]string{"alive": "true"})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Name string
	}{Name: "Test"}
	renderTemplate(w, "index", data)
}

func (h Handler) InfoIDHandler(w http.ResponseWriter, r *http.Request) {
	var entry = new(types.Entry)
	id := r.PathValue("imdb")
	if !validPath.MatchString(id) {
		http.Error(w, "not a valid id", http.StatusBadRequest)
		return
	}
	req, err := media.NewOmdbIDRequest(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mov, err := req.SendRequest()
	if err != nil && err.Error() == "response value is false" {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}
	entry.Movie = mov
	renderTemplate(w, "info", *entry)
}

// TODO: Cache Movie
func (h Handler) InfoTitleYearHandler(w http.ResponseWriter, r *http.Request) {
	var entry = new(types.Entry)
	title := r.PathValue("title")
	year := r.PathValue("year")
	req, err := media.NewOmdbTitleRequest(title, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mov, err := req.SendRequest()
	if err != nil {
		http.Error(w, fmt.Sprintf("error retrieving movie %s (%s)", title, year), http.StatusBadRequest)
		return
	}
	entry.Movie = mov
	renderTemplate(w, "info", *entry)
}

func (h Handler) CreateEntryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusBadRequest)
	}
	name := r.FormValue("name")
	watched := false
	if r.FormValue("watched") == "on" {
		watched = true
	}

	comment := r.FormValue("comment")
	id := r.PathValue("imdb")
	req, err := media.NewOmdbIDRequest(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	mov, err := req.SendRequest()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}
	entry := &types.Entry{
		ID:      0,
		Name:    name,
		Watched: watched,
		Comment: []byte(comment),
		Movie:   mov,
	}
	entry, err = h.store.CreateEntry(entry)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}
	renderTemplate(w, "info", *entry)
}

func renderTemplate(w http.ResponseWriter, tmpl string, d any) {
	err := templates.ExecuteTemplate(w, tmpl+".gohtml", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
