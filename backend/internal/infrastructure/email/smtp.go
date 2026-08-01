package email

import (
	"context"
	"fmt"
	"time"

	"github.com/go-mail/mail/v2"

	"github.com/muhammadyunus/Restify-Service/internal/config"
	domservice "github.com/muhammadyunus/Restify-Service/internal/domain/service"
)

// SMTPEmailService sends emails via SMTP.
type SMTPEmailService struct {
	host     string
	port     int
	username string
	password string
	from     string
}

// NewSMTPEmailService creates a new SMTP email service.
func NewSMTPEmailService(cfg config.SMTPConfig) domservice.EmailService {
	from := cfg.User
	if from == "" {
		from = "noreply@ForgeBase.api"
	}
	return &SMTPEmailService{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.User,
		password: cfg.Pass,
		from:     from,
	}
}

func (s *SMTPEmailService) send(ctx context.Context, to, subject, body string) error {
	d := mail.NewDialer(s.host, s.port, s.username, s.password)
	d.Timeout = 10 * time.Second

	msg := mail.NewMessage()
	msg.SetAddressHeader("From", s.from, "ForgeBase")
	msg.SetHeader("To", []string{to}...)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html; charset=UTF-8", body)

	if err := d.DialAndSend(msg); err != nil {
		return fmt.Errorf("send email: %w", err)
	}
	return nil
}

func (s *SMTPEmailService) SendAlertEmail(ctx context.Context, recipient, subject, body string) error {
	return s.send(ctx, recipient, subject, body)
}

func (s *SMTPEmailService) SendWelcomeEmail(ctx context.Context, recipient, name string) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<body>
    <h2>Welcome to ForgeBase, %s!</h2>
    <p>Your account has been created successfully. You can now start building your API backend.</p>
    <p>Best regards,<br>The ForgeBase Team</p>
</body>
</html>
`, name)
	return s.send(ctx, recipient, "Welcome to ForgeBase", htmlBody)
}

// Compile-time check.
var _ domservice.EmailService = (*SMTPEmailService)(nil)
