package handlers

import (
	"fmt"
	"github.com/jhachmer/gotocollection/internal/types"
	"log"
	"net/http"
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
