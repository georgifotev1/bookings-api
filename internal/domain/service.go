package domain

import "time"

type Service struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    int       `json:"duration"`
	Buffer      int       `json:"buffer"`
	Cost        int       `json:"cost"`
	Visible     bool      `json:"visible"`
	TenantID    string    `json:"tenant_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
