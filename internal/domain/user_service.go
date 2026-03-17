package domain

import "time"

type UserService struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ServiceID string    `json:"service_id"`
	TenantID  string    `json:"tenant_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
