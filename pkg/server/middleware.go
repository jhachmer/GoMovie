package server

import (
	"log"
	"net/http"
	"time"
)

type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc

// Chain is called with an HandlerFunc and one or more Middleware functions
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}

type LogMessage struct {
	Path     string
	Method   string
	Duration time.Duration
	Time     time.Time
}

func NewLogMessage(r *http.Request, startTime time.Time) *LogMessage {
	return &LogMessage{
		Path:     r.URL.EscapedPath(),
		Method:   r.Method,
		Duration: time.Since(startTime),
		Time:     time.Now(),
	}
}

func Logging(logger *log.Logger) Middleware {
	return func(handlerFunc http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				message := NewLogMessage(r, time.Now())
				logger.Println(message.Method, message.Path, message.Duration.String())
			}()
			handlerFunc(w, r)
		}
	}
}
