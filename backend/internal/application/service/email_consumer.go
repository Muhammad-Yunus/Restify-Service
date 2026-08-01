package service

import (
	"context"
	"encoding/json"
	"fmt"

	domservice "github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// EmailConsumer processes email messages from the queue.
type EmailConsumer struct {
	emailSvc domservice.EmailService
}

// NewEmailConsumer creates a new email consumer.
func NewEmailConsumer(emailSvc domservice.EmailService) *EmailConsumer {
	return &EmailConsumer{emailSvc: emailSvc}
}

// Handle processes an email notification message.
func (ec *EmailConsumer) Handle(ctx context.Context, message []byte) error {
	var msg struct {
		To       string `json:"to"`
		Subject  string `json:"subject"`
		Body     string `json:"body"`
		Template string `json:"template,omitempty"`
	}
	if err := json.Unmarshal(message, &msg); err != nil {
		return fmt.Errorf("parse email message: %w", err)
	}

	if err := ec.emailSvc.SendAlertEmail(ctx, msg.To, msg.Subject, msg.Body); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}
