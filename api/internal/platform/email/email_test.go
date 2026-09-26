package email

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
)

func validMessage() Message {
	return Message{To: "ana@exemplo.com", Subject: "Recuperação de acesso", TextBody: "Use o link: https://app.tj.example/redefinir?t=SEGREDO-DO-TOKEN"}
}

// EML-01.2
func TestValidateAcceptsAWellFormedMessage(t *testing.T) {
	if err := validMessage().Validate(); err != nil {
		t.Errorf("mensagem válida recusada: %v", err)
	}
}

func TestValidateRejectsInvalidRecipientsSubjectsAndBodies(t *testing.T) {
	cases := map[string]func(*Message){
		"destinatário vazio":              func(m *Message) { m.To = "" },
		"destinatário sem arroba":         func(m *Message) { m.To = "ana.exemplo.com" },
		"destinatário com nome":           func(m *Message) { m.To = "Ana <ana@exemplo.com>" },
		"destinatário duplo":              func(m *Message) { m.To = "a@x.com, b@x.com" },
		"quebra de linha no destinatário": func(m *Message) { m.To = "ana@exemplo.com\r\nBcc: alvo@x.com" },
		"LF no destinatário":              func(m *Message) { m.To = "ana@exemplo.com\nBcc: alvo@x.com" },
		"assunto vazio":                   func(m *Message) { m.Subject = "" },
		"assunto em branco":               func(m *Message) { m.Subject = "   " },
		"CR no assunto":                   func(m *Message) { m.Subject = "Olá\rBcc: alvo@x.com" },
		"LF no assunto":                   func(m *Message) { m.Subject = "Olá\nBcc: alvo@x.com" },
		"byte nulo no assunto":            func(m *Message) { m.Subject = "Olá\x00" },
		"corpo vazio":                     func(m *Message) { m.TextBody = "" },
		"destinatário grande demais":      func(m *Message) { m.To = strings.Repeat("a", 250) + "@x.com" },
		"assunto grande demais":           func(m *Message) { m.Subject = strings.Repeat("a", 999) },
	}
	for name, edit := range cases {
		m := validMessage()
		edit(&m)
		if err := m.Validate(); err == nil {
			t.Errorf("%s: deveria ser recusada", name)
		}
	}
}

func newLog(buf *bytes.Buffer) *slog.Logger { return logx.New("debug", buf) }

// EML-01.3: o remetente log registra só o domínio e o assunto.
func TestLogSenderLogsOnlyTheRecipientDomainAndTheSubject(t *testing.T) {
	var buf bytes.Buffer
	s, err := NewSender("log", newLog(&buf))
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Send(context.Background(), validMessage()); err != nil {
		t.Fatalf("Send: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "exemplo.com") || !strings.Contains(out, "Recuperação de acesso") {
		t.Errorf("o log deveria ter o domínio e o assunto: %s", out)
	}
	for _, leaked := range []string{"SEGREDO-DO-TOKEN", "https://app.tj.example", "ana@", "\"ana\""} {
		if strings.Contains(out, leaked) {
			t.Errorf("o log vazou %q: %s", leaked, out)
		}
	}
}

func TestLogSenderRejectsAnInvalidMessageWithoutLoggingIt(t *testing.T) {
	var buf bytes.Buffer
	s, _ := NewSender("log", newLog(&buf))
	m := validMessage()
	m.To = "inválido"

	if err := s.Send(context.Background(), m); err == nil {
		t.Error("esperava erro")
	}
	if strings.Contains(buf.String(), "Recuperação") {
		t.Errorf("uma mensagem recusada não deve ser registrada: %s", buf.String())
	}
}

func TestDisabledSenderSendsNothingAndWarnsAtCreation(t *testing.T) {
	var buf bytes.Buffer
	s, err := NewSender("disabled", newLog(&buf))
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(buf.String(), `"level":"WARN"`) || !strings.Contains(buf.String(), "desabilitado") {
		t.Errorf("deveria haver um aviso operacional na criação: %s", buf.String())
	}
	buf.Reset()
	if err := s.Send(context.Background(), validMessage()); err != nil {
		t.Errorf("Send com e-mail desabilitado: %v", err)
	}
	if strings.Contains(buf.String(), "exemplo.com") || strings.Contains(buf.String(), "SEGREDO") {
		t.Errorf("nada deveria ser registrado do conteúdo: %s", buf.String())
	}
}

func TestDisabledSenderStillValidatesMessagesToCatchBugsEarly(t *testing.T) {
	s, _ := NewSender("disabled", newLog(&bytes.Buffer{}))
	m := validMessage()
	m.Subject = ""

	if err := s.Send(context.Background(), m); err == nil {
		t.Error("uma mensagem inválida deveria ser recusada mesmo com o e-mail desabilitado")
	}
}

// EML-01.5
func TestNewSenderRejectsAnUnknownProviderNamingTheVariable(t *testing.T) {
	_, err := NewSender("smtp", newLog(&bytes.Buffer{}))

	if err == nil || !strings.Contains(err.Error(), "EMAIL_PROVIDER") {
		t.Errorf("err = %v", err)
	}
}
