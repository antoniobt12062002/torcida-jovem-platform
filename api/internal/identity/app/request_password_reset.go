package app

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/email"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
)

// Ports of the recovery use cases.
type (
	RecoveryTokens interface {
		RegisterRequest(ctx context.Context, emailHash []byte, at time.Time) (bool, error)
		CreateToken(ctx context.Context, userID string, tokenHash []byte, at, expiresAt time.Time, requestID string) error
	}
	EmailSender interface {
		Send(ctx context.Context, m email.Message) error
	}
)

const (
	sendTimeout  = 10 * time.Second
	resetSubject = "Recuperação de acesso à plataforma"
)

// RequestPasswordReset handles "I forgot my password". It answers the same
// whether or not the account exists (the handler sends the fixed message), so it
// returns an error only for internal failures.
//
// For an active account it creates a 256-bit token, stores only its SHA-256
// (valid for TTL, invalidating earlier pending tokens) and sends the link by
// e-mail after answering: the token travels in the URL fragment, so it does not
// reach access logs or Referer headers. At most three requests per e-mail per
// hour count; the others get the same answer without a token or an e-mail. The
// token, the link and the address never reach the audit trail or the logs.
type RequestPasswordReset struct {
	Users    UserFinder
	Recovery RecoveryTokens
	Sender   EmailSender
	Audit    SecurityAuditor
	Tx       TxFunc
	Now      Clock
	HashKey  []byte
	TTL      time.Duration
	BaseURL  string // without trailing slash, e.g. https://app.tj.example
	Log      *slog.Logger
	// Async runs the send after the answer; nil means a goroutine tracked by Wait.
	Async func(fn func())

	wg sync.WaitGroup
}

type RequestResetInput struct{ Email string }

// Wait blocks until the e-mails being sent in the background are done (tests and shutdown).
func (uc *RequestPasswordReset) Wait() { uc.wg.Wait() }

func (uc *RequestPasswordReset) ready() bool {
	return uc.Users != nil && uc.Recovery != nil && uc.Sender != nil && uc.Audit != nil && uc.Tx != nil && uc.Now != nil &&
		len(uc.HashKey) > 0 && uc.TTL > 0 && uc.BaseURL != "" && uc.Log != nil
}

func (uc *RequestPasswordReset) Execute(ctx context.Context, in RequestResetInput) error {
	if !uc.ready() {
		return errNotConfigured
	}
	now := uc.Now()
	emailHash, err := domain.HashIdentifier(uc.HashKey, in.Email)
	if err != nil {
		return err
	}
	hashHex := hex.EncodeToString(emailHash)

	allowed, err := uc.Recovery.RegisterRequest(ctx, emailHash, now)
	if err != nil {
		return err
	}
	result, outcome := "throttled", audit.OutcomeDenied
	entityType, entityID := "login", hashHex
	if allowed {
		result, outcome = "no_account", audit.OutcomeSuccess
		if normalized, nerr := domain.NormalizeEmail(in.Email); nerr == nil {
			user, ferr := uc.Users.FindByEmail(ctx, normalized)
			switch {
			case errors.Is(ferr, domain.ErrUserNotFound):
			case ferr != nil:
				return ferr
			case !user.Active:
				result, entityType, entityID = "inactive", "user", user.ID
			default:
				if err := uc.issue(ctx, user, now); err != nil {
					return err
				}
				result, entityType, entityID = "sent", "user", user.ID
			}
		}
	}
	uc.Audit.RecordSecurity(ctx, audit.Entry{
		ActorType: audit.ActorAnonymous, Action: audit.AuthPasswordResetRequested, EntityType: entityType, EntityID: entityID,
		Outcome: outcome, Context: map[string]any{"email_hash": hashHex, "result": result},
	})
	return nil
}

// issue creates the token and schedules the e-mail.
func (uc *RequestPasswordReset) issue(ctx context.Context, user domain.User, now time.Time) error {
	token, tokenHash, err := domain.NewSessionToken()
	if err != nil {
		return err
	}
	requestID := httpx.RequestIDFrom(ctx)
	err = uc.Tx(ctx, func(ctx context.Context) error {
		return uc.Recovery.CreateToken(ctx, user.ID, tokenHash, now, now.Add(uc.TTL), requestID)
	})
	if err != nil {
		return err
	}
	msg := email.Message{
		To:      user.Email,
		Subject: resetSubject,
		TextBody: fmt.Sprintf("Olá, %s.\n\nRecebemos um pedido para redefinir a senha da sua conta. Use o link abaixo em até %d minutos:\n\n%s/redefinir-senha#token=%s\n\nSe você não fez esse pedido, ignore esta mensagem: a sua senha continua a mesma.\n",
			strings.TrimSpace(user.Name), int(uc.TTL.Minutes()), uc.BaseURL, token),
	}
	send := func() {
		sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), sendTimeout)
		defer cancel()
		if err := uc.Sender.Send(sendCtx, msg); err != nil {
			uc.Log.Error("falha ao enviar o e-mail de recuperação de acesso",
				slog.String("incident", "password_reset_email_failed"),
				slog.String("request_id", requestID),
				slog.String("error", err.Error()))
		}
	}
	if uc.Async != nil {
		uc.Async(send)
		return nil
	}
	uc.wg.Go(send)
	return nil
}
