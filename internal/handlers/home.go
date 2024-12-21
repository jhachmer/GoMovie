package handlers

import (
	"github.com/jhachmer/gotocollection/internal/util"
	"net/http"
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
