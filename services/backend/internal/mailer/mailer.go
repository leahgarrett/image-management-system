package mailer

import (
	"fmt"
	"log"
	"net/smtp"
)

// Mailer sends transactional emails.
type Mailer interface {
	SendMagicLink(toEmail, magicLinkURL string) error
}

// SMTPMailer sends emails via SMTP.
type SMTPMailer struct {
	host string
	port int
	from string
}

// NewSMTPMailer creates an SMTPMailer.
func NewSMTPMailer(host string, port int, from string) *SMTPMailer {
	return &SMTPMailer{host: host, port: port, from: from}
}

// SendMagicLink sends the magic link email via SMTP.
func (m *SMTPMailer) SendMagicLink(to, url string) error {
	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	msg := []byte(fmt.Sprintf(
		"To: %s\r\nFrom: %s\r\nSubject: Your sign-in link\r\n\r\nClick to sign in: %s\r\n",
		to, m.from, url,
	))
	return smtp.SendMail(addr, nil, m.from, []string{to}, msg)
}

// LogMailer logs magic links instead of sending emails. Use in DEV_MODE.
type LogMailer struct{}

// NewLogMailer creates a LogMailer.
func NewLogMailer() *LogMailer { return &LogMailer{} }

// SendMagicLink logs the magic link URL to stdout.
func (m *LogMailer) SendMagicLink(to, url string) error {
	log.Printf("[DEV] Magic link for %s: %s", to, url)
	return nil
}
