package domain

import "time"

type Event struct {
	ID         string    `json:"id"`
	CustomerID string    `json:"customer_id"`
	ServiceID  string    `json:"service_id"`
	UserID     string    `json:"user_id"`
	TenantID   string    `json:"tenant_id"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Status     string    `json:"status"`
	Notes      string    `json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
