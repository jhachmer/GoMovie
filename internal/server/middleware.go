package server

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jhachmer/gomovie/internal/auth"
	"github.com/jhachmer/gomovie/internal/rate"
)

type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc

// Chain is called with an HandlerFunc and one or more Middleware functions
// wraps HandlerFunc f around one or more middleware functions
func Chain(handlerFunc http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		handlerFunc = m(handlerFunc)
	}
	return handlerFunc
}

// LogMessage holds data contained in log messages
type LogMessage struct {
	Path   string
	Method string
	Time   time.Time
}

// NewLogMessage returns pointer to a new LogMessage instance
func NewLogMessage(r *http.Request) *LogMessage {
	return &LogMessage{
		Path:   r.URL.EscapedPath(),
		Method: r.Method,
		Time:   time.Now(),
	}
}

// Logging is a middleware function that logs Path, Method, Duration
func Logging() Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				message := NewLogMessage(r)
				slog.Info("request", "method", message.Method, "path", message.Path, "took", time.Since(message.Time).String())
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
				slog.Warn("cookie missing", "err", err.Error())
				http.Redirect(w, r, "/login", http.StatusUnauthorized)
				return
			}
			_, err = auth.VerifyToken(cookie.Value)
			if err != nil {
				slog.Warn("jwt not verified", "err", err.Error())
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

func RateLimit(rl *rate.RateLimiter) Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			if idx := strings.LastIndex(ip, ":"); idx != -1 {
				ip = ip[:idx]
			}

			if !rl.Allow(ip) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			handlerFunc(w, r)
		}
	}
}
