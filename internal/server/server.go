package server

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/jhachmer/gotocollection/internal/handlers"
)

// Server struct with Address and Logger fields
type Server struct {
	Addr    string
	Logger  *log.Logger
	Handler *handlers.Handler
}

// NewServer returns a new Server instance with given Address and Logger and Handler values
func NewServer(addr string, logger *log.Logger, handler *handlers.Handler) *Server {
	svr := &Server{
		Addr:    addr,
		Logger:  logger,
		Handler: handler,
	}
	return svr
}

// setupRoutes initializes the URL Routes of the Server
// Handlers are wrapped with Middleware
func (svr *Server) setupRoutes(mux *http.ServeMux) {
	fileServer := http.FileServer(http.Dir("./templates/"))
	mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /health", Chain(svr.Handler.HealthHandler, Logging(svr.Logger)))
	mux.HandleFunc("GET /films/{imdb}", Chain(svr.Handler.InfoIDHandler, Logging(svr.Logger)))
	mux.HandleFunc("POST /films/{imdb}", Chain(svr.Handler.CreateEntryHandler, Logging(svr.Logger)))
	mux.HandleFunc("PUT /films/{imdb}", Chain(svr.Handler.UpdateEntryHandler, Logging(svr.Logger)))
	mux.HandleFunc("DELETE /films/{imdb}", Chain(svr.Handler.DeleteEntryHandler, Logging(svr.Logger)))

	mux.HandleFunc("GET /overview", Chain(svr.Handler.HomeHandler, Logging(svr.Logger)))
	mux.HandleFunc("GET /search", Chain(svr.Handler.SearchHandler, Logging(svr.Logger)))
}

// Serve calls setup functions and spins up the Server
func (svr *Server) Serve(ctx context.Context) error {
	handlers.InitTemplates()
	mux := http.NewServeMux()
	svr.setupRoutes(mux)

	server := &http.Server{
		Addr:    svr.Addr,
		Handler: mux,
	}

	errCh := make(chan error, 1)
	defer close(errCh)

	go func() {
		svr.Logger.Println("Starting server on " + svr.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		svr.Logger.Println("Shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			svr.Logger.Printf("Error during server shutdown: %v", err)
			return err
		}
		svr.Logger.Println("Server stopped")
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
