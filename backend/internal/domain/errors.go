package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound       = errors.New("resource not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrConflict       = errors.New("resource conflict")
	ErrInternal       = errors.New("internal server error")
	ErrServiceOffline = errors.New("service is unavailable")
)

type AppError struct {
	Err     error  `json:"-"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code int, message string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Err:     err,
	}
}
