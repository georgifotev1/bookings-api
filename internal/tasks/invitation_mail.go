package tasks

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/georgifotev1/nuvelaone-api/pkg/mailer"
	"github.com/hibiken/asynq"
	"go.uber.org/zap"
)

type InvitationEmailHandler struct {
	mailer mailer.Mailer
	logger *zap.SugaredLogger
}

func NewInvitationEmailHandler(m mailer.Mailer, logger *zap.SugaredLogger) *InvitationEmailHandler {
	return &InvitationEmailHandler{mailer: m, logger: logger}
}

func (h *InvitationEmailHandler) ProcessTask(ctx context.Context, t *asynq.Task) error {
	var p InvitationEmailPayload
	if err := json.Unmarshal(t.Payload(), &p); err != nil {
		return fmt.Errorf("InvitationEmailHandler: %w", err)
	}

	if h.mailer == nil {
		h.logger.Warnw("mailer not configured, skipping invitation email", "email", p.Email)
		return nil
	}

	h.logger.Infow("sending invitation email", "email", p.Email)

	if err := h.mailer.Send(mailer.EmailData{
		To:       []string{p.Email},
		Subject:  "You're invited to join " + p.TenantName,
		Template: "invitation",
		Data: map[string]any{
			"name":        p.Name,
			"invited_by":  p.InvitedByName,
			"tenant_name": p.TenantName,
			"accept_url":  p.AcceptURL,
		},
	}); err != nil {
		return fmt.Errorf("InvitationEmailHandler send: %w", err)
	}

	h.logger.Infow("invitation email sent", "email", p.Email)
	return nil
}
