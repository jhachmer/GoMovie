package handlers

import (
	"net/http"

	"github.com/jhachmer/gotocollection/internal/types"
)

type StatsPage struct {
	WatchStats *types.WatchStats
}

func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "error", http.StatusNotImplemented)
}
