package models

import (
	"database/sql"
	"github.com/google/uuid"
)

type Company struct {
	ID    uuid.UUID      `json:"id"`
	Name  string         `json:"name"`
	Image sql.NullString `json:"image,omitempty"`
}