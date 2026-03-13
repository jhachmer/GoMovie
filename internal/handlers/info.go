package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/jhachmer/gomovie/internal/api"
)

func (h *Handler) InfoIDHandler(w http.ResponseWriter, r *http.Request) {
	data := api.MovieInfoPage{}
	id := r.PathValue("imdb")
	if !validPath.MatchString(id) {
		http.Error(w, "not a valid id", http.StatusBadRequest)
		data.Error = fmt.Errorf("error validating imdb id: %s", id)
		slog.Error("could not match id", "id", id, "handler", "info_id", "err", data.Error.Error())
		return
	}
	mov, err := h.getMovie(id)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		data.Error = fmt.Errorf("error getting movie, %w", err)
		data.Movie = &api.Movie{}
		slog.Error("error getting movie", "handler", "info_id", "err", err.Error())
		renderTemplate(w, "info", data)
		return
	}
	data.Movie = mov
	entries, err := h.store.GetEntries(id)
	if err != nil {
		//http.Error(w, fmt.Sprintf("error getting movie %s", err.Error()), http.StatusInternalServerError)
		data.Error = fmt.Errorf("error getting entries")
		slog.Error("error getting entries", "handler", "info_id", "err", err.Error())
		renderTemplate(w, "info", data)
		return
	}
	data.Entries = entries
	renderTemplate(w, "info", data)
}

func (h *Handler) CreateMovieHandler(w http.ResponseWriter, r *http.Request) {
	data := api.MovieInfoPage{}
	id := r.PathValue("imdb")
	mov, err := h.getMovie(id)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = fmt.Errorf("error getting movie: %w", err)
		slog.Error("error getting movies", "handler", "create_movie", "err", err.Error())
		renderTemplate(w, "info", data)
	}
	_, err = h.store.CreateMovie(mov)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = fmt.Errorf("error saving movie: %w", err)
		slog.Error("error saving movie", "handler", "create_movie", "err", err.Error())
		renderTemplate(w, "info", data)
	}
	http.Redirect(w, r, fmt.Sprintf("/films/%s", id), http.StatusSeeOther)

}

func (h *Handler) ContainsMovieHandler(w http.ResponseWriter, r *http.Request) {
	imdbID := r.PathValue("imdb")
	_, err := h.store.GetMovieByID(imdbID)
	json.NewEncoder(w).Encode(map[string]bool{"exists": err == nil})
}

func (h *Handler) UpdateMovieHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("imdb")
	if !validPath.MatchString(id) {
		http.Error(w, "not a valid id", http.StatusBadRequest)
		slog.Error("could not match id", "id", id, "handler", "update_movie")
		return
	}
	updatedMovie, err := api.MovieFromID(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting movie: %s", err.Error()), http.StatusInternalServerError)
		slog.Error("error getting movie", "handler", "update_movie", "err", err.Error())
		return
	}
	_, err = h.store.UpdateMovie(updatedMovie)
	if err != nil {
		http.Error(w, fmt.Sprintf("error getting movie: %s", err.Error()), http.StatusInternalServerError)
		slog.Error("error updating movie", "handler", "update_movie", "err", err.Error())
		return
	}
	h.movCache.Delete(id)
	h.movCache.Set(id, updatedMovie)
}

func (h *Handler) DeleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("imdb")
	err := h.store.DeleteMedia(id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error deleting movie: %s", err.Error()), http.StatusInternalServerError)
		slog.Error("error deleting movie", "handler", "delete_movie", "err", err.Error())
		return
	}
	h.movCache.Delete(id)
}

func (h *Handler) CreateEntryHandler(w http.ResponseWriter, r *http.Request) {
	data := api.MovieInfoPage{}
	err := r.ParseForm()
	if err != nil {
		//http.Error(w, "error parsing form", http.StatusInternalServerError)
		data.Error = fmt.Errorf("error parsing form: %w", err)
		slog.Error("error parsing form", "handler", "create_movie", "err", err.Error())
		renderTemplate(w, "info", data)
	}
	name := r.FormValue("name")
	watched := r.FormValue("watched") == "on"

	comment := r.FormValue("comment")
	id := r.PathValue("imdb")
	mov, err := h.getMovie(id)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = fmt.Errorf("error getting movie: %w", err)
		slog.Error("error getting movie", "handler", "create_movie", "err", err.Error())
		renderTemplate(w, "info", data)
	}
	entry := api.NewEntry(name, watched, comment)
	_, err = h.store.CreateEntry(entry, mov)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = fmt.Errorf("error creating entry: %w", err)
		slog.Error("error creating entry", "handler", "create_movie", "err", err.Error())
		renderTemplate(w, "info", data)
	}
	entries, err := h.store.GetEntries(id)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = fmt.Errorf("error getting entries: %w", err)
		slog.Error("error getting entries", "handler", "create_movie", "err", err.Error())
		renderTemplate(w, "info", data)
	}
	data.Entries = entries
	data.Movie = mov
	http.Redirect(w, r, fmt.Sprintf("/films/%s", id), http.StatusSeeOther)
	//renderTemplate(w, "info", data)
}

func (h *Handler) UpdateEntryHandler(w http.ResponseWriter, r *http.Request) {
	movieId := r.PathValue("imdb")

	var payload struct {
		Name    string `json:"name"`
		Watched bool   `json:"watched"`
		Comment string `json:"comment"`
	}
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		slog.Error("error decoding payload", "handler", "update_entry", "err", err.Error())
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	_, err = h.store.UpdateEntry(movieId, payload.Name, payload.Comment, payload.Watched)
	if err != nil {
		slog.Error("error updating entry", "handler", "update_entry", "err", err.Error())
		http.Error(w, "error updating entry", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func (h *Handler) DeleteEntryHandler(w http.ResponseWriter, r *http.Request) {
	movieId := r.PathValue("imdb")
	err := h.store.DeleteEntry(movieId)
	if err != nil {
		slog.Error("error deleting entry", "handler", "delete_entry", "err", err.Error())
		http.Error(w, "error deleting entry", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
