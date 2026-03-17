package domain

import "time"

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	AddressID string    `json:"address_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WorkingHours struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	DayOfWeek int    `json:"day_of_week"`
	OpensAt   string `json:"opens_at"`
	ClosesAt  string `json:"closes_at"`
	IsClosed  bool   `json:"is_closed"`
}
