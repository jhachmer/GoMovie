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
	Mux     *http.ServeMux
}

// NewServer returns a new Server instance with given Address and Logger and Handler values
func NewServer(addr string, logger *log.Logger, handler *handlers.Handler) *Server {
	mux := http.NewServeMux()
	svr := &Server{
		Addr:    addr,
		Logger:  logger,
		Handler: handler,
		Mux:     mux,
	}
	return svr
}

// setupRoutes initializes the URL Routes of the Server
// Handlers are wrapped with Middleware
func (svr *Server) setupRoutes() {
	fileServer := http.FileServer(http.Dir("./templates/"))

	svr.Mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	svr.Mux.HandleFunc("GET /health", Chain(svr.Handler.HealthHandler, Logging(svr.Logger)))
	svr.Mux.Handle("GET /{$}", http.RedirectHandler("/login", http.StatusSeeOther))
	svr.Mux.HandleFunc("GET /login", Chain(svr.Handler.LoginHandler, RedirectWhenLoggedIn(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("POST /login", Chain(svr.Handler.CheckLoginHandler, RedirectWhenLoggedIn(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("GET /register", Chain(svr.Handler.RegisterSiteHandler, RedirectWhenLoggedIn(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("POST /register", Chain(svr.Handler.RegisterHandler, RedirectWhenLoggedIn(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("GET /films/{imdb}", Chain(svr.Handler.InfoIDHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("PUT /films/{imdb}", Chain(svr.Handler.UpdateMovieHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("DELETE /films/{imdb}", Chain(svr.Handler.DeleteMovieHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("POST /films/{imdb}/entry", Chain(svr.Handler.CreateEntryHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("PUT /films/{imdb}/entry", Chain(svr.Handler.UpdateEntryHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("DELETE /films/{imdb}/entry", Chain(svr.Handler.DeleteEntryHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("GET /overview", Chain(svr.Handler.HomeHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("GET /search", Chain(svr.Handler.SearchHandler, Authenticate(), Logging(svr.Logger)))
	svr.Mux.HandleFunc("GET /stats", Chain(svr.Handler.StatsHandler, Authenticate(), Logging(svr.Logger)))
}

// Serve calls setup functions and spins up the Server
func (svr *Server) Serve(ctx context.Context) error {
	handlers.InitTemplates()
	svr.setupRoutes()

	server := &http.Server{
		Addr:    svr.Addr,
		Handler: svr.Mux,
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
