package domain

import "github.com/google/uuid"

// NewUUID generates a new UUID v4.
func NewUUID() uuid.UUID {
	return uuid.New()
}
