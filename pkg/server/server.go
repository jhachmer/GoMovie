package server

import (
	"log"
	"net/http"
)

// Server struct with Address and Logger fields
type Server struct {
	Addr    string
	Logger  *log.Logger
	Handler *Handler
}

// NewServer returns a new Server instance with given Address and Logger values
func NewServer(addr string, logger *log.Logger, handler *Handler) *Server {
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
	mux.HandleFunc("GET /films/{title}/{year}", Chain(svr.Handler.InfoTitleYearHandler, Logging(svr.Logger)))
	mux.HandleFunc("GET /{$}", Chain(svr.Handler.HomeHandler, Logging(svr.Logger)))

	mux.HandleFunc("POST /films/{imdb}", Chain(svr.Handler.CreateEntryHandler, Logging(svr.Logger)))
}

// Serve calls setup functions and spins up the Server
func (svr *Server) Serve() {
	mux := http.NewServeMux()
	svr.setupRoutes(mux)
	svr.Logger.Println("Starting server on " + svr.Addr)
	log.Fatal(http.ListenAndServe(svr.Addr, mux))
}
