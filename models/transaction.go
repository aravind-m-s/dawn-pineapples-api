package models

import (
	"time"

	"github.com/google/uuid"
)

type Transaction struct {
	ID        uuid.UUID `json:"id"`
	Date      time.Time `json:"date"`
	KG        float64   `json:"kg"`
	Rate      float64   `json:"rate"`
	Amount    float64   `json:"amount"`
	Taxi      float64   `json:"taxi"`
	Cash      float64   `json:"cash"`
	Balance   float64   `json:"balance"`
	CompanyID uuid.UUID `json:"company_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}