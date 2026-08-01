# Epic 26 — Email Service

**Goal:** Implement email sending via SMTP for alerts, welcome emails, and notifications.
**Dependencies:** Epic 06 (Logger adapter, RabbitMQ adapter), Epic 21 (MessageQueue interface)
**Commit:** `feat: add SMTP email service for notifications`

---

## Step 26.01 — Email Service Implementation

**Build:** Create `backend/internal/infrastructure/email/smtp.go`:

```go
package email

import (
    "context"
    "fmt"
    "strings"
    "time"

    "github.com/go-mail/mail/v2"
    "github.com/muhammadyunus/ForgeBase/internal/config"
    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
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
func NewSMTPEmailService(cfg config.SMTPConfig) repository.EmailService {
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
```

---

## Step 26.02 — Email Queue Consumer

**Build:** Create `backend/internal/application/service/email_consumer.go`:

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/muhammadyunus/ForgeBase/internal/domain/repository"
)

// EmailConsumer processes email messages from the queue.
type EmailConsumer struct {
    emailSvc repository.EmailService
}

// NewEmailConsumer creates a new email consumer.
func NewEmailConsumer(emailSvc repository.EmailService) *EmailConsumer {
    return &EmailConsumer{emailSvc: emailSvc}
}

// Handle processes an email notification message.
func (ec *EmailConsumer) Handle(ctx context.Context, message []byte) error {
    var msg struct {
        To        string `json:"to"`
        Subject   string `json:"subject"`
        Body      string `json:"body"`
        Template  string `json:"template,omitempty"`
    }
    if err := json.Unmarshal(message, &msg); err != nil {
        return fmt.Errorf("parse email message: %w", err)
    }

    if err := ec.emailSvc.SendAlertEmail(ctx, msg.To, msg.Subject, msg.Body); err != nil {
        return fmt.Errorf("send email: %w", err)
    }
    return nil
}
```

---

## Step 26.03 — Register Email Worker

**Build:** Update `internal/di/bootstrap.go`:

```go
// In bootstrap, register email consumer
emailConsumer := email.NewEmailConsumer(c.EmailService)
queue.Consume(ctx, "email.notifications", emailConsumer.Handle)
```

**Test cases:**
- [ ] Unit: `SendWelcomeEmail()` formats email correctly
- [ ] Unit: `SendAlertEmail()` sends via SMTP
- [ ] Integration: Full email send cycle with test SMTP server

---

## Commit Instruction

```bash
git add .
git commit -m "feat: add SMTP email service with welcome and alert emails"
```
