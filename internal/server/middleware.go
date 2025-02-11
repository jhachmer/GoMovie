package server

import (
	"log"
	"net/http"
	"time"

	"github.com/jhachmer/gomovie/internal/auth"
)

type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc

// Chain is called with an HandlerFunc and one or more Middleware functions
// wraps HandlerFunc f around one or more middleware functions
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

// LogMessage holds data contained in log messages
type LogMessage struct {
	Path   string
	Method string
	Time   time.Time
}

// NewLogMessage returns pointer to a new LogMessage instance
func NewLogMessage(r *http.Request, startTime time.Time) *LogMessage {
	return &LogMessage{
		Path:   r.URL.EscapedPath(),
		Method: r.Method,
		Time:   time.Now(),
	}
}

// Logging is a middleware function that logs Path, Method, Duration
func Logging(logger *log.Logger) Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				message := NewLogMessage(r, time.Now())
				logger.Println(message.Method, message.Path, time.Since(message.Time).String())
			}()
			handlerFunc(w, r)
		}
	}
}

func Authenticate() Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("gomovie")
			if err != nil {
				log.Println("cookie missing", err)
				http.Redirect(w, r, "/login", http.StatusUnauthorized)
				return
			}
			_, err = auth.VerifyToken(cookie.Value)
			if err != nil {
				log.Println("jwt not verified", err)
				http.Redirect(w, r, "/login", http.StatusUnauthorized)
				return
			}
			handlerFunc(w, r)
		}
	}
}

func RedirectWhenLoggedIn() Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("gomovie")
			if err == nil {
				_, err = auth.VerifyToken(cookie.Value)
				if err == nil {
					http.Redirect(w, r, "/overview", http.StatusSeeOther)
					return
				}
			}

			handlerFunc(w, r)
		}
	}
}
