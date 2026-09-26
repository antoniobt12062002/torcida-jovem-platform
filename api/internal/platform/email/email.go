// Package email is the platform's e-mail capability: a sending interface that
// does not depend on any provider. Modules use Sender only; a real provider
// (Resend, SES, SendGrid, SMTP) is added later as another implementation.
package email

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
)

const (
	maxAddressLen = 254
	maxSubjectLen = 200
)

// Message is a plain-text e-mail to a single recipient.
type Message struct {
	To       string
	Subject  string
	TextBody string
}

// Validate rejects what could corrupt the headers or that no provider accepts:
// a recipient that is not a single bare address, an empty or oversized subject,
// CR, LF or NUL in the recipient or the subject (header injection), and an empty
// body.
func (m Message) Validate() error {
	if m.To == "" || len(m.To) > maxAddressLen || hasControl(m.To) {
		return errors.New("email: destinatário inválido")
	}
	addr, err := mail.ParseAddress(m.To)
	if err != nil || addr.Address != m.To {
		return errors.New("email: o destinatário deve ser um único endereço, sem nome")
	}
	if strings.TrimSpace(m.Subject) == "" || len(m.Subject) > maxSubjectLen || hasControl(m.Subject) {
		return errors.New("email: assunto inválido")
	}
	if strings.TrimSpace(m.TextBody) == "" {
		return errors.New("email: corpo vazio")
	}
	return nil
}

func hasControl(s string) bool { return strings.ContainsAny(s, "\r\n\x00") }

// Sender sends messages. Implementations must call Message.Validate first.
type Sender interface {
	Send(ctx context.Context, m Message) error
}

// NewSender builds the sender for EMAIL_PROVIDER: "log" (development: logs the
// recipient domain and the subject, never the body) or "disabled" (sends
// nothing). Real providers will be added here.
func NewSender(provider string, log *slog.Logger) (Sender, error) {
	switch provider {
	case "log":
		return logSender{log: log}, nil
	case "disabled":
		log.Warn("e-mail desabilitado (EMAIL_PROVIDER=disabled): nenhuma mensagem será enviada")
		return disabledSender{}, nil
	default:
		return nil, fmt.Errorf("EMAIL_PROVIDER inválido: %q (use log ou disabled)", provider)
	}
}

type logSender struct{ log *slog.Logger }

// Send logs only the recipient domain and the subject: the body may carry a
// recovery link, and the local part of the address is personal data.
func (s logSender) Send(ctx context.Context, m Message) error {
	if err := m.Validate(); err != nil {
		return err
	}
	s.log.InfoContext(ctx, "e-mail simulado (provedor log)",
		slog.String("recipient_domain", m.To[strings.LastIndexByte(m.To, '@')+1:]),
		slog.String("subject", m.Subject))
	return nil
}

type disabledSender struct{}

func (disabledSender) Send(_ context.Context, m Message) error { return m.Validate() }
