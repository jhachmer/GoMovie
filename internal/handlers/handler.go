package handlers

import (
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/jhachmer/gomovie/internal/api"
	"github.com/jhachmer/gomovie/internal/cache"
	"github.com/jhachmer/gomovie/internal/store"
)

var validPath = regexp.MustCompile(`^tt\d{7,8}$`)

// var validYear = regexp.MustCompile(`^(19|20)\d{2}$`)

type Handler struct {
	store    store.Store
	movCache *cache.Cache[string, *api.Movie]
	serCache *cache.Cache[string, *api.Series]
}

func NewHandler(store store.Store, movC *cache.Cache[string, *api.Movie], serC *cache.Cache[string, *api.Series]) *Handler {
	return &Handler{
		store:    store,
		movCache: movC,
		serCache: serC,
	}
}

func (h *Handler) Close() {
	h.movCache.Close()
	h.serCache.Close()
	h.store.Close()
}

func (h *Handler) getMovie(id string) (*api.Movie, error) {
	if mov, ok := h.movCache.Get(id); ok {
		slog.Info("found media in cache", "id", id)
		return mov, nil
	}
	if mov, err := h.store.GetMovieByID(id); err == nil {
		slog.Info("found media in db", "id", id)
		h.movCache.Set(id, mov)
		return mov, nil
	}
	if mov, err := api.MovieFromID(id); err == nil {
		slog.Info("got media from api", "id", id)
		h.movCache.Set(id, mov)
		return mov, nil
	}
	return nil, fmt.Errorf("error getting movie with id: %s", id)
}

var templates *template.Template

func InitTemplates() {
	funcMap := template.FuncMap{"perc": perc}
	templates = template.Must(template.New("gomovie").Funcs(funcMap).ParseFiles(
		"./templates/index.html",
		"./templates/info.html",
		"./templates/overview.html",
		"./templates/movie-grid.html",
		"./templates/error.html",
		"./templates/register.html",
		"./templates/admin.html",
		"./templates/stats.html"))
}

func perc(num1, num2 int) float32 {
	return (float32(num1) / float32(num2)) * 100
}

func renderTemplate(w http.ResponseWriter, tmpl string, d any) {
	err := templates.ExecuteTemplate(w, tmpl+".html", d)
	if err != nil {
		slog.Error("error rendering template", "err", err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
