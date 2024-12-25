package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jhachmer/gotocollection/internal/types"
)

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
	page := types.InfoPage{
		Entries: entries,
		Movie:   mov,
	}
	renderTemplate(w, "info", page)
}

func (h *Handler) InfoTitleYearHandler(w http.ResponseWriter, r *http.Request) {
	var page = new(types.InfoPage)
	title := r.PathValue("title")
	year := r.PathValue("year")
	if !validYear.MatchString(year) {
		http.Error(w, "not a valid year", http.StatusBadRequest)
		log.Println("could not match year", year)
		return
	}
	mov, err := types.NewMovieFromTitleAndYear(title, year)
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
	entry := types.NewEntry(name, watched, comment)
	_, err = h.store.CreateEntry(entry, mov)
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
	page := types.NewInfoPage(mov, entries)
	renderTemplate(w, "info", page)
}
