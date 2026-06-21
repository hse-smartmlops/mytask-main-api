// Package mail is the outbound SMTP adapter (implements ports.Mailer).
// net/smtp stays confined here; nothing above the adapter layer imports it.
package mail

import (
	"fmt"
	"net/smtp"
	"os"
	"strings"

	"emplacc-api/internal/ports"
)

type smtpMailer struct {
	addr     string // host:port
	host     string
	from     string
	username string
	password string
}

// New builds the SMTP mailer from SMTP_* env vars. Returns nil (no-op mailer)
// when SMTP_HOST is unset, so notifications degrade gracefully without email.
func New() ports.Mailer {
	host := os.Getenv("SMTP_HOST")
	if host == "" {
		return nil
	}
	port := os.Getenv("SMTP_PORT")
	if port == "" {
		port = "587"
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = os.Getenv("SMTP_USERNAME")
	}
	return &smtpMailer{
		addr:     host + ":" + port,
		host:     host,
		from:     from,
		username: os.Getenv("SMTP_USERNAME"),
		password: os.Getenv("SMTP_PASSWORD"),
	}
}

func (m *smtpMailer) Send(to, subject, body string) error {
	if to == "" {
		return fmt.Errorf("empty recipient")
	}
	var auth smtp.Auth
	if m.username != "" {
		auth = smtp.PlainAuth("", m.username, m.password, m.host)
	}
	msg := strings.Join([]string{
		"From: " + m.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")
	return smtp.SendMail(m.addr, auth, m.from, []string{to}, []byte(msg))
}
