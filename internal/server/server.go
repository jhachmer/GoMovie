package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/jhachmer/gomovie/internal/handlers"
	"github.com/jhachmer/gomovie/internal/rate"
)

// Server struct with Address and Logger fields
type Server struct {
	Addr        string
	Handler     *handlers.Handler
	Mux         *http.ServeMux
	RateLimiter *rate.RateLimiter
}

// NewServer returns a new Server instance with given Address and Logger and Handler values
func NewServer(addr string, handler *handlers.Handler) *Server {
	mux := http.NewServeMux()

	rateLimiter := rate.NewRateLimiter(100, time.Minute)
	svr := &Server{
		Addr:        addr,
		Handler:     handler,
		Mux:         mux,
		RateLimiter: rateLimiter,
	}
	return svr
}

// setupRoutes initializes the URL Routes of the Server
// Handlers are wrapped with Middleware
func (svr *Server) setupRoutes() {
	fileServer := http.FileServer(http.Dir("./templates/"))

	svr.Mux.Handle("GET /static/", http.StripPrefix("/static", fileServer))
	svr.Mux.HandleFunc("GET /health", Chain(svr.Handler.HealthHandler, Logging()))
	svr.Mux.Handle("GET /{$}", http.RedirectHandler("/login", http.StatusSeeOther))
	svr.Mux.HandleFunc("GET /login", Chain(svr.Handler.LoginHandler, RedirectWhenLoggedIn(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("POST /login", Chain(svr.Handler.CheckLoginHandler, RedirectWhenLoggedIn(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("GET /register", Chain(svr.Handler.RegisterSiteHandler, RedirectWhenLoggedIn(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("POST /register", Chain(svr.Handler.RegisterHandler, RedirectWhenLoggedIn(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("GET /films/{imdb}", Chain(svr.Handler.InfoIDHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("POST /films/{imdb}", Chain(svr.Handler.CreateMovieHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("PUT /films/{imdb}", Chain(svr.Handler.UpdateMovieHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("DELETE /films/{imdb}", Chain(svr.Handler.DeleteMovieHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("POST /films/{imdb}/entry", Chain(svr.Handler.CreateEntryHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("PUT /films/{imdb}/entry", Chain(svr.Handler.UpdateEntryHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("DELETE /films/{imdb}/entry", Chain(svr.Handler.DeleteEntryHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("GET /overview", Chain(svr.Handler.HomeHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("GET /search", Chain(svr.Handler.SearchHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("GET /stats", Chain(svr.Handler.StatsHandler, Authenticate(), RateLimit(svr.RateLimiter), Logging()))
	svr.Mux.HandleFunc("GET /check/{imdb}", Chain(svr.Handler.ContainsMovieHandler, RateLimit(svr.RateLimiter), Authenticate(), Logging()))

	svr.Mux.HandleFunc("GET /admin", Chain(svr.Handler.AdminHandler, Logging()))
	svr.Mux.HandleFunc("POST /admin_login", Chain(svr.Handler.AdminLoginHandler, Logging()))
	svr.Mux.HandleFunc("GET /get_users", Chain(svr.Handler.GetUsersHandler, Logging()))
	svr.Mux.HandleFunc("PUT /toggle_active", Chain(svr.Handler.ToggleActiveHandler, Logging()))
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
		slog.Info("server started", "addr", svr.Addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		slog.Info("Shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("Error during server shutdown", "err", err)
			return err
		}
		slog.Error("Server stopped")
		return ctx.Err()
	case err := <-errCh:
		return err
	}
}
