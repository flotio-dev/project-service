package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error       string `json:"error"`
	Description string `json:"description,omitempty"`
}

type SuccessResponse[T any] struct {
	Data T `json:"data"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("httpx: encode error: %v", err)
	}
}

// 2xx
func OK[T any](w http.ResponseWriter, v T) { writeJSON(w, http.StatusOK, SuccessResponse[T]{Data: v}) }
func Created[T any](w http.ResponseWriter, v T) {
	writeJSON(w, http.StatusCreated, SuccessResponse[T]{Data: v})
}
func NoContent(w http.ResponseWriter) { writeJSON(w, http.StatusNoContent, nil) }

// 3xx convenience (rare as JSON)
func RedirectJSON(w http.ResponseWriter, location string, permanent bool) {
	status := http.StatusFound
	if permanent {
		status = http.StatusMovedPermanently
	}
	writeJSON(w, status, map[string]string{"location": location})
}

// 4xx
func BadRequest(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusBadRequest, ErrorResponse{Error: "bad_request", Description: msg})
}
func Unauthorized(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusUnauthorized, ErrorResponse{Error: "unauthorized", Description: msg})
}
func Forbidden(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusForbidden, ErrorResponse{Error: "forbidden", Description: msg})
}
func NotFound(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusNotFound, ErrorResponse{Error: "not_found", Description: msg})
}
func Conflict(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusConflict, ErrorResponse{Error: "conflict", Description: msg})
}

// 5xx
func InternalError(w http.ResponseWriter, msg string) {
	writeJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Description: msg})
}
