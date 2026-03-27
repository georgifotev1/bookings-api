package domain

import (
	"strings"
	"time"
	"unicode"

	"github.com/georgifotev1/nuvelaone-api/pkg/timeutil"
)

type Tenant struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email"`
	Tier      Tier      `json:"tier"`
	Timezone  string    `json:"timezone"`
	AddressID *string   `json:"address_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Tier string

const (
	TierBase Tier = "base"
	TierPro  Tier = "pro"
)

type WorkingHours struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	DayOfWeek int    `json:"day_of_week"`
	OpensAt   string `json:"opens_at"`
	ClosesAt  string `json:"closes_at"`
	IsClosed  bool   `json:"is_closed"`
}

type WorkingHoursRequest struct {
	DayOfWeek int    `json:"day_of_week"`
	OpensAt   string `json:"opens_at"`
	ClosesAt  string `json:"closes_at"`
	IsClosed  bool   `json:"is_closed"`
}

type UpdateTenantRequest struct {
	Name         string                `json:"name" validate:"omitempty,min=1,max=50"`
	Phone        string                `json:"phone" validate:"omitempty,max=20"`
	Email        string                `json:"email" validate:"omitempty,email"`
	Timezone     string                `json:"timezone" validate:"omitempty,timezone"`
	WorkingHours []WorkingHoursRequest `json:"working_hours" validate:"omitempty,dive"`
}

func NewSlug(name string) string {
	var result strings.Builder
	for _, r := range strings.ToLower(name) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
		} else if r == ' ' || r == '-' || r == '_' {
			result.WriteRune('-')
		}
	}
	return result.String()
}

func ParseTimeInLocation(timeStr string, loc *time.Location) (time.Time, error) {
	today := time.Now().Format(timeutil.DateOnly + " ")
	return time.ParseInLocation(timeutil.DateTime, today+timeStr, loc)
}
