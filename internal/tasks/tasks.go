package tasks

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
)

const (
	TaskMaxRetry = 5
	TaskTimeout  = 30 * time.Second

	TypeWelcomeEmail         = "email:welcome"
	TypeInvitationEmail      = "email:invitation"
	TypeCleanupExpiredTokens = "token:cleanup"
)

type WelcomeEmailPayload struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}

func NewWelcomeEmailTask(p WelcomeEmailPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("tasks.NewWelcomeEmailTask: %w", err)
	}
	return asynq.NewTask(TypeWelcomeEmail, payload,
		asynq.MaxRetry(TaskMaxRetry),
		asynq.Timeout(TaskTimeout),
	), nil
}

type InvitationEmailPayload struct {
	InvitedByName string `json:"invited_by_name"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Token         string `json:"token"`
	TenantName    string `json:"tenant_name"`
	AcceptURL     string `json:"accept_url"`
}

func NewInvitationEmailTask(p InvitationEmailPayload) (*asynq.Task, error) {
	payload, err := json.Marshal(p)
	if err != nil {
		return nil, fmt.Errorf("tasks.NewInvitationEmailTask: %w", err)
	}
	return asynq.NewTask(TypeInvitationEmail, payload,
		asynq.MaxRetry(TaskMaxRetry),
		asynq.Timeout(TaskTimeout),
	), nil
}

func NewCleanupExpiredTokensTask() *asynq.Task {
	return asynq.NewTask(TypeCleanupExpiredTokens, nil,
		asynq.MaxRetry(1),
		asynq.Timeout(60*time.Second),
	)
}
