package handlers

import (
	"fmt"
	"github.com/jhachmer/gotocollection/internal/cache"
	"github.com/jhachmer/gotocollection/internal/store"
	"github.com/jhachmer/gotocollection/internal/types"
	"log"
	"net/http"
	"regexp"
)

var validPath = regexp.MustCompile("^tt\\d{7,8}$")
var validYear = regexp.MustCompile("^(19|20)\\d{2}$")

type Handler struct {
	store    store.Store
	movCache *cache.Cache[string, *types.Movie]
}

func NewHandler(store store.Store, movC *cache.Cache[string, *types.Movie]) *Handler {
	return &Handler{
		store:    store,
		movCache: movC,
	}
}

func (h *Handler) Close() {
	h.movCache.Close()
}

func (h *Handler) getMovie(id string) (*types.Movie, error) {
	if mov, ok := h.movCache.Get(id); ok {
		log.Printf("got movie with id %s from cache", id)
		return mov, nil
	}
	if mov, err := h.store.GetMovie(id); err == nil {
		log.Printf("got movie with id %s from db", id)
		h.movCache.Set(id, mov)
		return mov, nil
	}
	if mov, err := types.NewMovieFromID(id); err == nil {
		log.Printf("got movie with id %s from api", id)
		h.movCache.Set(id, mov)
		return mov, nil
	}
	return nil, fmt.Errorf("could not get movie")
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
	page := types.NewInfoPage(mov, entries)
	renderTemplate(w, "info", page)
}
