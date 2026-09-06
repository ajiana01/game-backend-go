package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/ajiana01/game-backend-go/internal/domainerr"
)

// ErrorResponse is the standard API error payload.
type ErrorResponse struct {
	Error string `json:"error"`
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		WriteError(w, http.StatusBadRequest, "invalid JSON body")
		return false
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteError(w, http.StatusBadRequest, "request body must contain one JSON object")
		return false
	}
	return true
}

func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, ErrorResponse{Error: message})
}

func WriteServiceError(w http.ResponseWriter, err error) {
	var validation domainerr.ValidationError
	switch {
	case errors.As(err, &validation):
		WriteError(w, http.StatusUnprocessableEntity, validation.Error())
	case errors.Is(err, domainerr.ErrNotFound):
		WriteError(w, http.StatusNotFound, "resource not found")
	case errors.Is(err, domainerr.ErrConflict):
		WriteError(w, http.StatusConflict, "resource already exists")
	default:
		WriteError(w, http.StatusInternalServerError, "internal server error")
	}
}
