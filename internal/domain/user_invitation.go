package domain

import "time"

type UserInvitation struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	Phone     string    `json:"phone"`
	Token     string    `json:"token"`
	Role      string    `json:"role"`
	InvitedBy string    `json:"invited_by"`
	ExpiresAt time.Time `json:"expires_at"`
	TenantID  string    `json:"tenant_id"`
	Accepted  bool      `json:"accepted"`
	CreatedAt time.Time `json:"created_at"`
}
