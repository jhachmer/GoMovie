package handlers

import (
	"fmt"
	"log"
	"regexp"

	"github.com/jhachmer/gotocollection/internal/cache"
	"github.com/jhachmer/gotocollection/internal/store"
	"github.com/jhachmer/gotocollection/internal/types"
)

var validPath = regexp.MustCompile(`^tt\d{7,8}$`)
var validYear = regexp.MustCompile(`^(19|20)\d{2}$`)

type Handler struct {
	logger   *log.Logger
	store    store.Store
	movCache *cache.Cache[string, *types.Movie]
}

func NewHandler(store store.Store, movC *cache.Cache[string, *types.Movie], logger *log.Logger) *Handler {
	return &Handler{
		store:    store,
		movCache: movC,
		logger:   logger,
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
