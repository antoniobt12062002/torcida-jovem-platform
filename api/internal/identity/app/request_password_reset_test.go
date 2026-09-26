//go:build integration

package app_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email/emailtest"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
)

type resetEnv struct {
	*ucEnv
	mail *emailtest.Recorder
	uc   *app.RequestPasswordReset
}

func newResetEnv(t *testing.T) *resetEnv {
	t.Helper()
	e := newUCEnv(t)
	mail := &emailtest.Recorder{}
	return &resetEnv{ucEnv: e, mail: mail, uc: &app.RequestPasswordReset{
		Users: e.users, Recovery: e.recovery, Sender: mail, Audit: e.rec, Tx: e.tx, Now: e.clock,
		HashKey: hashKey, TTL: 30 * time.Minute, BaseURL: "https://app.tj.example", Log: logx.New("info", e.logs),
		Async: func(fn func()) { fn() },
	}}
}

// tokenFromMail extracts the token from the recovery link in the last message.
func (r *resetEnv) tokenFromMail(t *testing.T) string {
	t.Helper()
	msg, ok := r.mail.Last()
	if !ok {
		t.Fatal("nenhum e-mail enviado")
	}
	_, after, found := strings.Cut(msg.TextBody, "https://app.tj.example/redefinir-senha#token=")
	if !found {
		t.Fatalf("o corpo deveria ter o link com o token no fragmento: %q", msg.TextBody)
	}
	return strings.Fields(after)[0]
}

func (r *resetEnv) request(email string) error {
	return r.uc.Execute(r.ctx, app.RequestResetInput{Email: email})
}

// IDN-07.1 e .2: token de 256 bits, só o hash no banco, e-mail com o link.
func TestRequestCreatesATokenStoresOnlyItsHashAndSendsTheLink(t *testing.T) {
	r := newResetEnv(t)
	u, _ := r.actor(t)

	if err := r.request("  " + strings.ToUpper(u.Email) + " "); err != nil {
		t.Fatal(err)
	}

	msgs := r.mail.Messages()
	if len(msgs) != 1 || msgs[0].To != u.Email || msgs[0].Subject != "Recuperação de acesso à plataforma" || !strings.Contains(msgs[0].TextBody, "30 minutos") {
		t.Fatalf("mensagens = %+v", msgs)
	}
	token := r.tokenFromMail(t)
	if len(token) != 43 {
		t.Errorf("o token deveria ter 256 bits em base64url (43 caracteres), tem %d", len(token))
	}
	got, err := r.recovery.FindValid(context.Background(), sha256of(token), r.now.Add(29*time.Minute))
	if err != nil || got != u.ID {
		t.Errorf("FindValid = %q, %v", got, err)
	}
	if _, err := r.recovery.FindValid(context.Background(), sha256of(token), r.now.Add(30*time.Minute)); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("o token vale 30 minutos: %v", err)
	}
	var stored string
	_ = r.owner.Raw("SELECT row_to_json(p)::text FROM password_reset_tokens p").Scan(&stored).Error
	if strings.Contains(stored, token) || !strings.Contains(stored, "req-12345678") {
		t.Errorf("linha do token = %s", stored)
	}
}

// IDN-07.3: nem token, nem link, nem e-mail em log ou auditoria.
func TestNothingSensitiveGoesToTheAuditTrailOrTheLogs(t *testing.T) {
	r := newResetEnv(t)
	u, _ := r.actor(t)
	_ = r.request(u.Email)
	token := r.tokenFromMail(t)

	evs := r.events(t, "auth.password_reset_requested")
	if len(evs) != 1 || evs[0].ActorType != "anonymous" || evs[0].EntityID != u.ID || evs[0].Outcome != "success" {
		t.Fatalf("eventos = %+v", evs)
	}
	blob := evs[0].Context + r.logs.String()
	for _, secret := range []string{token, "redefinir-senha", u.Email, "@exemplo.com"} {
		if strings.Contains(blob, secret) {
			t.Errorf("vazou %q: %s", secret, blob)
		}
	}
	if !strings.Contains(evs[0].Context, "email_hash") {
		t.Errorf("contexto = %s", evs[0].Context)
	}
}

// IDN-07.1: o resultado não revela se a conta existe.
func TestUnknownInactiveAndInvalidEmailsReturnTheSameResultAndSendNothing(t *testing.T) {
	r := newResetEnv(t)
	off, _ := r.actor(t)
	_ = r.users.SetActive(context.Background(), off.ID, false)

	for name, addr := range map[string]string{"inexistente": "ninguem@exemplo.com", "inativo": off.Email, "inválido": "isto não é um e-mail"} {
		if err := r.request(addr); err != nil {
			t.Errorf("%s: o resultado deveria ser igual ao de uma conta existente, veio %v", name, err)
		}
	}

	if len(r.mail.Messages()) != 0 || r.count(t, "password_reset_tokens") != 0 {
		t.Error("nenhum token nem e-mail para quem não pode recuperar")
	}
	if n := len(r.events(t, "auth.password_reset_requested")); n != 3 {
		t.Errorf("as três solicitações são auditadas: %d", n)
	}
}

// IDN-07.4: no máximo 3 por e-mail por hora, com o mesmo resultado.
func TestMoreThanThreeRequestsPerHourCreateNoTokenAndSendNoEmail(t *testing.T) {
	r := newResetEnv(t)
	u, _ := r.actor(t)

	for i := range 3 {
		r.now = t0.Add(time.Duration(i) * time.Minute)
		if err := r.request(u.Email); err != nil {
			t.Fatal(err)
		}
	}
	firstThree := len(r.mail.Messages())
	r.now = t0.Add(10 * time.Minute)
	if err := r.request(u.Email); err != nil {
		t.Errorf("a quarta tem o mesmo resultado: %v", err)
	}

	if firstThree != 3 || len(r.mail.Messages()) != 3 {
		t.Errorf("e-mails = %d/%d, esperado 3", firstThree, len(r.mail.Messages()))
	}
	evs := r.events(t, "auth.password_reset_requested")
	if len(evs) != 4 || evs[3].Outcome != "denied" || !strings.Contains(evs[3].Context, "throttled") {
		t.Errorf("a quarta deveria ser auditada como negada: %+v", evs)
	}
	r.now = t0.Add(61 * time.Minute)
	if err := r.request(u.Email); err != nil || len(r.mail.Messages()) != 4 {
		t.Errorf("depois de uma hora volta a valer: %v, e-mails = %d", err, len(r.mail.Messages()))
	}
}

// IDN-07.2: um novo pedido invalida o token anterior.
func TestANewRequestInvalidatesThePreviousToken(t *testing.T) {
	r := newResetEnv(t)
	u, _ := r.actor(t)
	_ = r.request(u.Email)
	first := r.tokenFromMail(t)
	r.now = t0.Add(time.Minute)

	_ = r.request(u.Email)
	second := r.tokenFromMail(t)

	if _, err := r.recovery.FindValid(context.Background(), sha256of(first), r.now); !errors.Is(err, domain.ErrInvalidResetToken) {
		t.Errorf("o primeiro token deveria ter sido invalidado: %v", err)
	}
	if _, err := r.recovery.FindValid(context.Background(), sha256of(second), r.now); err != nil {
		t.Errorf("o segundo vale: %v", err)
	}
}

// EML-01.4: falha de envio é incidente no log e não chega ao cliente.
func TestEmailDeliveryFailureIsAnIncidentAndDoesNotChangeTheResult(t *testing.T) {
	r := newResetEnv(t)
	u, _ := r.actor(t)
	r.mail.FailWith = errors.New("provedor indisponível")

	if err := r.request(u.Email); err != nil {
		t.Errorf("a falha de envio não pode aparecer para o cliente: %v", err)
	}

	out := r.logs.String()
	if !strings.Contains(out, "password_reset_email_failed") || !strings.Contains(out, "req-12345678") || strings.Contains(out, u.Email) || strings.Contains(out, "redefinir-senha") {
		t.Errorf("o incidente deveria estar no log, sem dados sensíveis: %s", out)
	}
}

// O envio acontece depois da resposta: por padrão é assíncrono.
func TestSendingIsAsynchronousByDefault(t *testing.T) {
	r := newResetEnv(t)
	r.uc.Async = nil
	u, _ := r.actor(t)

	if err := r.request(u.Email); err != nil {
		t.Fatal(err)
	}
	r.uc.Wait()

	if len(r.mail.Messages()) != 1 {
		t.Errorf("e-mails = %d", len(r.mail.Messages()))
	}
}

func TestRequestRequiresAllDependencies(t *testing.T) {
	if err := (&app.RequestPasswordReset{}).Execute(context.Background(), app.RequestResetInput{}); err == nil {
		t.Error("um caso de uso mal montado deveria falhar")
	}
}

// blockingSender holds every send until released and records the context state.
type blockingSender struct {
	release chan struct{}
	ctxErr  chan error
}

func (b *blockingSender) Send(ctx context.Context, _ email.Message) error {
	<-b.release
	b.ctxErr <- ctx.Err()
	return nil
}

// A resposta não espera o envio, e o envio não é cancelado com a requisição.
func TestTheResponseDoesNotWaitForTheSendAndTheSendSurvivesTheRequestContext(t *testing.T) {
	r := newResetEnv(t)
	r.uc.Async = nil
	sender := &blockingSender{release: make(chan struct{}), ctxErr: make(chan error, 1)}
	r.uc.Sender = sender
	u, _ := r.actor(t)
	reqCtx, cancel := context.WithCancel(r.ctx)

	done := make(chan error, 1)
	go func() { done <- r.uc.Execute(reqCtx, app.RequestResetInput{Email: u.Email}) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("a resposta ficou esperando o envio")
	}
	cancel() // a requisição terminou
	close(sender.release)
	r.uc.Wait()

	if err := <-sender.ctxErr; err != nil {
		t.Errorf("o envio não pode ser cancelado quando a requisição termina: %v", err)
	}
}

type failingUsers struct{}

func (failingUsers) FindByEmail(context.Context, string) (domain.User, error) {
	return domain.User{}, errors.New("banco indisponível")
}

func TestAnInternalFailureIsReturnedAsAnError(t *testing.T) {
	r := newResetEnv(t)
	r.uc.Users = failingUsers{}

	if err := r.request("ana@exemplo.com"); err == nil {
		t.Error("uma falha interna não pode ser escondida como se a conta não existisse")
	}
}
