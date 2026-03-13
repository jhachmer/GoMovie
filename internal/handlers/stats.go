package handlers

import (
	"fmt"
	"net/http"

	"github.com/jhachmer/gomovie/internal/api"
)

type StatsPage struct {
	WatchStats *api.WatchStats
	Error      error
}

func newStatsPage(h *Handler) (StatsPage, error) {
	watchStats, err := h.store.GetWatchCounts()
	if err != nil {
		return StatsPage{}, err
	}
	return StatsPage{
		WatchStats: watchStats,
		Error:      nil,
	}, nil
}

func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	statsPage, err := newStatsPage(h)
	if err != nil {
		statsPage = StatsPage{
			Error: fmt.Errorf("could not fetch statistics add some movies first: %w", err),
		}
		renderTemplate(w, "stats", statsPage)
		return
	}
	renderTemplate(w, "stats", statsPage)
}
