package domain

import "time"

type UserInvitation struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Token     string    `json:"-"`
	Role      string    `json:"role"`
	InvitedBy string    `json:"invited_by"`
	ExpiresAt time.Time `json:"expires_at"`
	TenantID  string    `json:"tenant_id"`
	Accepted  bool      `json:"accepted"`
	CreatedAt time.Time `json:"created_at"`
}
