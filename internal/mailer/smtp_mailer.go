package mailer

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"

	"robot-hand-server/internal/domain"
)

type SMTPMailer struct {
	host                 string
	port                 int
	username             string
	password             string
	from                 string
	supportRecipientMail string
}

func NewSMTPMailer(host string, port int, username, password, from, supportRecipientMail string) (*SMTPMailer, error) {
	if strings.TrimSpace(host) == "" || port <= 0 || strings.TrimSpace(from) == "" {
		return nil, fmt.Errorf("smtp host, port, from are required")
	}

	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("smtp credentials are required")
	}

	return &SMTPMailer{
		host:                 host,
		port:                 port,
		username:             username,
		password:             password,
		from:                 from,
		supportRecipientMail: strings.TrimSpace(supportRecipientMail),
	}, nil
}

func (m *SMTPMailer) SendPasswordResetCode(_ context.Context, email, code string) error {
	subject := "Password Reset Code"
	body := fmt.Sprintf("Your password reset code is: %s\n\nCode expires in 15 minutes.", code)
	return m.send([]string{email}, subject, body)
}

func (m *SMTPMailer) SendSupportTicket(_ context.Context, ticket domain.SupportTicket) error {
	if m.supportRecipientMail == "" {
		return nil
	}

	subject := fmt.Sprintf("Support ticket %s", ticket.ID)
	body := fmt.Sprintf(
		"User UUID: %s\nEmail: %s\nHeader: %s\nText:\n%s\nImage: %s\nCreatedAt: %s\n",
		ticket.UserUUID,
		ticket.Email,
		ticket.Header,
		ticket.Text,
		ticket.ImagePath,
		ticket.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	)

	return m.send([]string{m.supportRecipientMail}, subject, body)
}

func (m *SMTPMailer) send(to []string, subject, body string) error {
	message := strings.Join([]string{
		fmt.Sprintf("From: %s", m.from),
		fmt.Sprintf("To: %s", strings.Join(to, ",")),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%d", m.host, m.port)
	if m.port == 465 {
		return m.sendImplicitTLS(addr, to, []byte(message))
	}
	return m.sendWithSTARTTLS(addr, to, []byte(message))
}

func (m *SMTPMailer) sendWithSTARTTLS(addr string, to []string, message []byte) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}

	if err := m.authenticate(client); err != nil {
		return err
	}

	return sendMessage(client, m.from, to, message)
}

func (m *SMTPMailer) sendImplicitTLS(addr string, to []string, message []byte) error {
	conn, err := tls.Dial("tcp", addr, &tls.Config{ServerName: m.host, MinVersion: tls.VersionTLS12})
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, m.host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()

	if err := m.authenticate(client); err != nil {
		return err
	}

	return sendMessage(client, m.from, to, message)
}

func (m *SMTPMailer) authenticate(client *smtp.Client) error {
	if ok, _ := client.Extension("AUTH"); !ok {
		return nil
	}
	return client.Auth(smtp.PlainAuth("", m.username, m.password, m.host))
}

func sendMessage(client *smtp.Client, from string, to []string, message []byte) error {
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}

	if _, err := writer.Write(message); err != nil {
		_ = writer.Close()
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	if err := client.Quit(); err != nil && !isNetworkClosed(err) {
		return err
	}
	return nil
}

func isNetworkClosed(err error) bool {
	return err != nil &&
		(errors.Is(err, net.ErrClosed) || strings.Contains(strings.ToLower(err.Error()), "closed network connection"))
}
