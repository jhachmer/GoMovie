package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/jhachmer/gotocollection/internal/types"
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

// TODO: invalid string handling (maybe regex?)
// genre:horror,thriller;actors:Hans Albers, Keeanu Reeves
func parseSearchQuery(query string) (types.SearchParams, error) {
	var sp types.SearchParams
	subQueries := strings.Split(query, ";")
	for i := range subQueries {
		subQueries[i] = strings.TrimSpace(subQueries[i])
	}
	for _, subquery := range subQueries {
		q := strings.Split(subquery, ":")
		if len(q) != 2 {
			return sp, fmt.Errorf("invalid search string passed %s", query)
		}
		searchType := strings.ToLower(q[0])
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
