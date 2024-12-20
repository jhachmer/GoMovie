package server

import (
	"fmt"
	"github.com/jhachmer/gotocollection/internal/cache"
	"github.com/jhachmer/gotocollection/internal/media"
	"github.com/jhachmer/gotocollection/internal/store"
	"github.com/jhachmer/gotocollection/internal/util"
	"html/template"
	"log"
	"net/http"
	"regexp"
)

var templates = template.Must(template.ParseFiles("./templates/index.gohtml", "./templates/info.gohtml"))
var validPath = regexp.MustCompile("^tt\\d{7,8}$")
var validYear = regexp.MustCompile("^(19|20)\\d{2}$")

type Handler struct {
	store    *store.Storage
	movCache *cache.Cache[string, *media.Movie]
}

func NewHandler(store *store.Storage, movC *cache.Cache[string, *media.Movie]) *Handler {
	return &Handler{
		store:    store,
		movCache: movC,
	}
}

func (h *Handler) Close() {
	h.movCache.Close()
}

func (h *Handler) getMovie(id string) (*media.Movie, error) {
	if mov, ok := h.movCache.Get(id); ok {
		log.Printf("got movie with id %s from cache", id)
		return mov, nil
	}
	if mov, err := h.store.GetMovie(id); err == nil {
		log.Printf("got movie with id %s from db", id)
		h.movCache.Set(id, mov)
		return mov, nil
	}
	if mov, err := media.NewMovieFromID(id); err == nil {
		log.Printf("got movie with id %s from api", id)
		h.movCache.Set(id, mov)
		return mov, nil
	}
	return nil, fmt.Errorf("could not get movie")
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := util.Encode(w, r, http.StatusOK, "healthy")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Name string
	}{Name: "Test"}
	renderTemplate(w, "index", data)
}

func (h *Handler) InfoIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("imdb")
	if !validPath.MatchString(id) {
		http.Error(w, "not a valid id", http.StatusBadRequest)
		log.Println("could not match id", id)
		return
	}
	mov, err := h.getMovie(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err.Error())
		return
	}
	entries, err := h.store.GetEntries(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting movie %s", err.Error()), http.StatusInternalServerError)
		log.Println(err.Error())
	}
	page := media.InfoPage{
		Entries: entries,
		Movie:   mov,
	}
	renderTemplate(w, "info", page)
}

func (h *Handler) InfoTitleYearHandler(w http.ResponseWriter, r *http.Request) {
	var page = new(media.InfoPage)
	title := r.PathValue("title")
	year := r.PathValue("year")
	if !validYear.MatchString(year) {
		http.Error(w, "not a valid year", http.StatusBadRequest)
		log.Println("could not match year", year)
		return
	}
	mov, err := media.NewMovieFromTitleAndYear(title, year)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting movie %s", err.Error()), http.StatusInternalServerError)
		log.Println(err.Error())
	}
	page.Movie = mov
	renderTemplate(w, "info", *page)
}

func (h *Handler) CreateEntryHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusInternalServerError)
		log.Println(err.Error())
	}
	name := r.FormValue("name")
	watched := r.FormValue("watched") == "on"

	comment := r.FormValue("comment")
	id := r.PathValue("imdb")
	mov, err := h.getMovie(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	entry := media.NewEntry(name, watched, comment)
	entry, err = h.store.CreateEntry(entry, mov)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	entries, err := h.store.GetEntries(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}
	page := media.NewInfoPage(mov, entries)
	renderTemplate(w, "info", page)
}

func renderTemplate(w http.ResponseWriter, tmpl string, d any) {
	err := templates.ExecuteTemplate(w, tmpl+".gohtml", d)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Println(err.Error())
	}
}
