package response

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/faultline/faultline/internal/domain"
)

// JSON writes a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

// Error writes an error JSON response.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, APIResponse{
		Error:   message,
		Success: false,
	})
}

// APIResponse is the standard response wrapper.
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Success bool        `json:"success"`
}

// Success writes a successful API response.
func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, APIResponse{
		Data:    data,
		Success: true,
	})
}

// Created writes a 201 response.
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, APIResponse{
		Data:    data,
		Success: true,
	})
}

// NotFound writes a 404 response.
func NotFound(w http.ResponseWriter, message string) {
	JSON(w, http.StatusNotFound, APIResponse{
		Error:   message,
		Success: false,
	})
}

// BadRequest writes a 400 response.
func BadRequest(w http.ResponseWriter, message string) {
	JSON(w, http.StatusBadRequest, APIResponse{
		Error:   message,
		Success: false,
	})
}

// InternalError writes a 500 response.
func InternalError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, APIResponse{
		Error:   message,
		Success: false,
	})
}

// HandleError maps domain errors to appropriate HTTP responses.
func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		JSON(w, appErr.Code, APIResponse{
			Error:   appErr.Message,
			Success: false,
		})
		return
	}

	switch {
	case errors.Is(err, domain.ErrNotFound):
		NotFound(w, err.Error())
	case errors.Is(err, domain.ErrInvalidInput):
		BadRequest(w, err.Error())
	case errors.Is(err, domain.ErrConflict):
		JSON(w, http.StatusConflict, APIResponse{
			Error:   err.Error(),
			Success: false,
		})
	default:
		InternalError(w, "internal server error")
	}
}
