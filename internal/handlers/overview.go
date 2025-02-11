package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jhachmer/gomovie/internal/types"
	"github.com/jhachmer/gomovie/internal/util"
)

// HealthHandler handles requestes to /health route
// Returns healthy as a JSON object if server is running
func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	err := util.Encode(w, r, http.StatusOK, "healthy")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// HomeHandler handles requests to /overview route
// lists all movies retrieved from database on the overview page
// movies retrieved get sorted before being passed to tempalte render
func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	data := types.HomeData{}
	movies, err := h.store.GetAllMovies()
	types.SortMovieSlice(movies)
	data.Movies = movies
	if err != nil {
		//http.Error(w, "error getting movies", http.StatusInternalServerError)
		h.logger.Println("HomeHandler: Getting Movies:", err)
		data.Error = err
	}
	renderTemplate(w, "overview", data)
}

// SearchHandler handles requests to /search route
// template html must have form with "query" input field
// input strings gets parsed into SearchParams type by parseSearchQuery function
// SearchParams are used in DB query
//
// Allowed search types are: Genre, Actors and Year
// Different search types must be separated by a semi-colon
// Search values are separated from the search type by colons
// Example string:
// genre:horror,thriller;actors:Hans Albers, Keeanu Reeves
func (h *Handler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	data := types.HomeData{}
	err := r.ParseForm()
	if err != nil {
		//http.Error(w, "error parsing form", http.StatusInternalServerError)
		data.Error = fmt.Errorf("error parsing form: %w", err)
		h.logger.Println(err.Error())
		renderTemplate(w, "overview", data)
	}
	query := r.FormValue("query")
	sp, err := parseSearchQuery(query)
	if err != nil {
		data.Error = fmt.Errorf("error parsing search query: %w", err)
		h.logger.Println(err.Error())
		renderTemplate(w, "overview", data)
		return
	}
	movs, err := h.store.SearchMovie(sp)
	if err != nil {
		//http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		data.Error = fmt.Errorf("error searching for movie: %w", err)
		h.logger.Println(err.Error())
		renderTemplate(w, "overview", data)
		return
	}
	data.Movies = movs
	renderTemplate(w, "overview", data)
}

// parseSearchQuery evaluates the search input for validity
// Splits String according to the rules and builds SearchParams instance
// which gets based to the database query
//
// Allowed search types are: Genre, Actors and Year
// Different search types must be separated by a semi-colon
// Search values are separated from the search type by colons
// Example string:
// genre:horror,thriller;actors:Hans Albers, Keeanu Reeves
func parseSearchQuery(query string) (types.SearchParams, error) {
	var sp types.SearchParams
	if query == "" {
		return sp, fmt.Errorf("search query must not be empty")
	}
	subQueries := strings.Split(query, ";")
	for i := range subQueries {
		subQueries[i] = strings.TrimSpace(subQueries[i])
	}
	for _, subquery := range subQueries {
		q := strings.Split(subquery, ":")
		if len(q) != 2 {
			return sp, fmt.Errorf("invalid search string passed: %s", query)
		}
		searchType := strings.ToLower(strings.TrimSpace(q[0]))
		values := q[1]
		switch searchType {
		case "genre":
			subVals := strings.Split(values, ",")
			for i := range subVals {
				subVals[i] = strings.TrimSpace(subVals[i])
			}
			sp.Genres = subVals
		case "actors":
			subVals := strings.Split(values, ",")
			for i := range subVals {
				subVals[i] = strings.TrimSpace(subVals[i])
			}
			sp.Actors = subVals
		case "year":
			yearParams := strings.Split(values, ",")
			sp.Years = types.YearSearch{StartYear: yearParams[0], EndYear: yearParams[1]}
		default:
			return sp, fmt.Errorf("invalid search type: %s", searchType)
		}
	}
	return sp, nil
}
