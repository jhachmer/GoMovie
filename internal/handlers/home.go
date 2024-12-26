package handlers

import (
	"log"
	"net/http"
	"strings"

	"github.com/jhachmer/gotocollection/internal/store"
	"github.com/jhachmer/gotocollection/internal/util"
)

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := util.Encode(w, r, http.StatusOK, "healthy")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	movies, err := h.store.GetAllMovies()
	if err != nil {
		http.Error(w, "error getting movies", http.StatusInternalServerError)
	}
	renderTemplate(w, "overview", movies)
}

func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "error parsing form", http.StatusInternalServerError)
		log.Println(err.Error())
	}
	query := r.FormValue("query")
	sp := parseSearchQuery(query)
	movs, err := h.store.SearchMovie(sp)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Println(err.Error())
	}
	log.Println(movs)
	renderTemplate(w, "overview", movs)
}

// TODOS: Spaces need to be handled further
// genre:horror,thriller;actors:Hans Albers, Keeanu Reeves
func parseSearchQuery(query string) store.SearchParams {
	var sp store.SearchParams
	subQueries := strings.Split(query, ";")
	for _, subquery := range subQueries {
		q := strings.Split(subquery, ":")
		searchType := q[0]
		values := q[1]
		switch searchType {
		case "genre":
			sp.Genres = strings.Split(values, ",")
		case "actors":
			sp.Actors = strings.Split(values, ",")
		case "year":
			yearParams := strings.Split(values, ",")
			sp.Years = store.YearSearch{StartYear: yearParams[0], EndYear: yearParams[1]}
		}
	}
	return sp
}
