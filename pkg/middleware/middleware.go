package middleware

import (
	"context"
	"errors"
	"net/http"
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
