package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"
)

// Chain permet de composer plusieurs middlewares ensemble.
func Chain(h http.Handler, m ...func(http.Handler) http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}
	return h
}

type contextKey string

const (
	ctxKeyToken  contextKey = "token"
	ctxKeyClaims contextKey = "claims"
	ctxKeySub    contextKey = "sub"
)

// WithValue ajoute une valeur dans le contexte.
func WithValue[T any](r *http.Request, key contextKey, val T) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, val))
}

// GetValue récupère une valeur typée depuis le contexte.
func GetValue[T any](r *http.Request, key contextKey) (T, bool) {
	v, ok := r.Context().Value(key).(T)
	return v, ok
}

// ErrUnauthorized est retournée quand la vérification d'auth échoue.
var ErrUnauthorized = errors.New("unauthorized")

// LoggingMiddleware loggue les requêtes HTTP entrantes (méthode, chemin, remote addr, durée, status)
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		// response status capture
		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)
		dur := time.Since(start)
		log.Printf("%s %s %s status=%d dur=%s", r.Method, r.RequestURI, r.RemoteAddr, rw.status, dur)
	})
}

// statusRecorder capture le status écrit par le handler
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// wrappers pour testabilité (remplacent time.Now/Since si besoin)
// no wrappers needed; using time.Now and time.Since directly
