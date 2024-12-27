package handlers

import (
	"fmt"
	"net/http"

	"github.com/jhachmer/gotocollection/internal/types"
)

func (h *Handler) InfoIDHandler(w http.ResponseWriter, r *http.Request) {
	data := types.InfoPage{}
	id := r.PathValue("imdb")
	if !validPath.MatchString(id) {
		//http.Error(w, "not a valid id", http.StatusBadRequest)
		data.Error = fmt.Errorf("error validating imdb id: %s", id)
		h.logger.Println("could not match id", id)
		renderTemplate(w, "info", data)
	}
	mov, err := h.getMovie(id)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusBadRequest)
		data.Error = err
		h.logger.Println(err.Error())
		renderTemplate(w, "info", data)
	}
	entries, err := h.store.GetEntries(id)
	if err != nil {
		//http.Error(w, fmt.Sprintf("error getting movie %s", err.Error()), http.StatusInternalServerError)
		data.Error = err
		h.logger.Println(err.Error())
		renderTemplate(w, "info", data)
	}
	data.Entries = entries
	data.Movie = mov
	renderTemplate(w, "info", data)
}

func (h *Handler) CreateEntryHandler(w http.ResponseWriter, r *http.Request) {
	data := types.InfoPage{}
	err := r.ParseForm()
	if err != nil {
		//http.Error(w, "error parsing form", http.StatusInternalServerError)
		data.Error = err
		h.logger.Println(err.Error())
		renderTemplate(w, "info", data)
	}
	name := r.FormValue("name")
	watched := r.FormValue("watched") == "on"

	comment := r.FormValue("comment")
	id := r.PathValue("imdb")
	mov, err := h.getMovie(id)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = err
		h.logger.Println(err.Error())
		renderTemplate(w, "info", data)
	}
	entry := types.NewEntry(name, watched, comment)
	_, err = h.store.CreateEntry(entry, mov)
	if err != nil {
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = err
		h.logger.Println(err.Error())
		renderTemplate(w, "info", data)
	}
	entries, err := h.store.GetEntries(id)
	if err != nil {
		//http.Error(w, err.Error(), http.StatusInternalServerError)
		data.Error = err
		h.logger.Println(err.Error())
		renderTemplate(w, "info", data)
	}
	data.Entries = entries
	data.Movie = mov
	renderTemplate(w, "info", data)
}
