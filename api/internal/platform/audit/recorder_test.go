//go:build integration

package audit_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/audit"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/httpx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/logx"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

const actorID = "0f8fad5b-d9cb-469f-a165-70867728950e"

type row struct {
	ActorType   string
	ActorUserID *string
	Action      string
	EntityType  string
	EntityID    string
	Before      *string
	After       *string
	Outcome     string
	Reason      *string
	Context     string
	RequestID   *string
}

type env struct {
	app, owner *gorm.DB
	rec        *audit.Recorder
	logs       *bytes.Buffer
}

func newEnv(t *testing.T) env {
	t.Helper()
	app := testutil.NewTestDB(t)
	logs := &bytes.Buffer{}
	return env{app: app, owner: testutil.OwnerDBFor(t, app), rec: audit.NewRecorder(app, logx.New("info", logs)), logs: logs}
}

func (e env) rows(t *testing.T) []row {
	t.Helper()
	var rows []row
	err := e.owner.Raw(`SELECT actor_type, actor_user_id::text AS actor_user_id, action, entity_type, entity_id,
		before::text AS before, after::text AS after, outcome, reason, context::text AS context, request_id
		FROM audit_log ORDER BY occurred_at, id`).Scan(&rows).Error
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	return rows
}

func ctxWith(requestID, actor string) context.Context {
	ctx := context.Background()
	if requestID != "" {
		ctx = httpx.WithRequestID(ctx, requestID)
	}
	if actor != "" {
		ctx = audit.WithActor(ctx, actor)
	}
	return ctx
}

func promotion() audit.Entry {
	return audit.Entry{
		Action: audit.AdminPromote, EntityType: "user", EntityID: "u-2", Outcome: audit.OutcomeSuccess,
		Reason: "eleito tesoureiro",
		Before: map[string]any{"roles": []any{"ASSOCIADO"}},
		After:  map[string]any{"roles": []any{"TESOURARIA"}, "password_hash": "SEGREDO"},
	}
}

// AUD-01.6: fora de uma transação, Record devolve erro e não grava.
func TestRecordOutsideATransactionFailsAndWritesNothing(t *testing.T) {
	e := newEnv(t)

	err := e.rec.Record(ctxWith("req-12345678", actorID), promotion())

	if !errors.Is(err, audit.ErrNoTransaction) {
		t.Errorf("err = %v, esperava ErrNoTransaction", err)
	}
	if n := len(e.rows(t)); n != 0 {
		t.Errorf("nada deveria ter sido gravado, linhas = %d", n)
	}
}

// AUD-01.1 e AUD-01.3: grava na transação do contexto, com request_id e ator do contexto.
func TestRecordWritesTheEntryWithContextActorAndRequestID(t *testing.T) {
	e := newEnv(t)

	err := database.WithTx(ctxWith("req-12345678", actorID), e.app, func(ctx context.Context) error {
		return e.rec.Record(ctx, promotion())
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	rows := e.rows(t)
	if len(rows) != 1 {
		t.Fatalf("linhas = %d", len(rows))
	}
	r := rows[0]
	if r.ActorType != "user" || r.ActorUserID == nil || *r.ActorUserID != actorID {
		t.Errorf("ator = %s/%v", r.ActorType, r.ActorUserID)
	}
	if r.RequestID == nil || *r.RequestID != "req-12345678" {
		t.Errorf("request_id = %v", r.RequestID)
	}
	if r.Action != "admin.promote" || r.EntityType != "user" || r.EntityID != "u-2" || r.Outcome != "success" ||
		r.Reason == nil || *r.Reason != "eleito tesoureiro" {
		t.Errorf("registro = %+v", r)
	}
	if r.Before == nil || !strings.Contains(*r.Before, "ASSOCIADO") || r.After == nil || !strings.Contains(*r.After, "TESOURARIA") {
		t.Errorf("before/after = %v / %v", r.Before, r.After)
	}
	if strings.Contains(*r.After, "SEGREDO") || !strings.Contains(*r.After, "[redacted]") {
		t.Errorf("o segredo deveria estar redigido: %s", *r.After)
	}
	if r.Context != "{}" {
		t.Errorf("context = %s", r.Context)
	}
}

func TestRecordKeepsAnExplicitActorAndStoresNullForMissingSnapshots(t *testing.T) {
	e := newEnv(t)
	entry := audit.Entry{
		ActorType: audit.ActorSystem, Action: audit.RBACSync, EntityType: "rbac", EntityID: "catalog",
		Outcome: audit.OutcomeSuccess, Context: map[string]any{"added": 2},
	}

	err := database.WithTx(ctxWith("req-12345678", actorID), e.app, func(ctx context.Context) error {
		return e.rec.Record(ctx, entry)
	})
	if err != nil {
		t.Fatalf("Record: %v", err)
	}

	r := e.rows(t)[0]
	if r.ActorType != "system" || r.ActorUserID != nil || r.Before != nil || r.After != nil {
		t.Errorf("registro = %+v", r)
	}
	var ctxJSON map[string]any
	if err := json.Unmarshal([]byte(r.Context), &ctxJSON); err != nil || ctxJSON["added"] != float64(2) {
		t.Errorf("context = %s (%v)", r.Context, err)
	}
}

func TestRecordRejectsAnInvalidEntryWithoutWriting(t *testing.T) {
	e := newEnv(t)
	bad := promotion()
	bad.Reason = ""

	err := database.WithTx(ctxWith("req-12345678", actorID), e.app, func(ctx context.Context) error {
		return e.rec.Record(ctx, bad)
	})

	if !errors.Is(err, audit.ErrInvalidEntry) || len(e.rows(t)) != 0 {
		t.Errorf("err = %v, linhas = %d", err, len(e.rows(t)))
	}
}

func TestRecordWithoutAnyActorIsRejected(t *testing.T) {
	e := newEnv(t)

	err := database.WithTx(ctxWith("req-12345678", ""), e.app, func(ctx context.Context) error {
		return e.rec.Record(ctx, promotion())
	})

	if !errors.Is(err, audit.ErrInvalidEntry) {
		t.Errorf("err = %v, esperava ErrInvalidEntry (sem ator)", err)
	}
}

// AUD-01.2: falha ao gravar devolve erro mapeável para 500 audit_failed.
func TestRecordFailureReturnsErrWrite(t *testing.T) {
	e := newEnv(t)
	if err := e.owner.Exec("ALTER TABLE audit_log RENAME TO audit_log_off").Error; err != nil {
		t.Fatal(err)
	}

	err := database.WithTx(ctxWith("req-12345678", actorID), e.app, func(ctx context.Context) error {
		return e.rec.Record(ctx, promotion())
	})

	if !errors.Is(err, audit.ErrWrite) {
		t.Errorf("err = %v, esperava ErrWrite", err)
	}
}

// AUD-04.6: RecordSecurity grava em transação própria, sem depender da de negócio.
func TestRecordSecuritySurvivesTheRollbackOfTheBusinessTransaction(t *testing.T) {
	e := newEnv(t)
	denied := audit.Entry{
		ActorType: audit.ActorUser, ActorID: actorID, Action: audit.RoleChangeDenied,
		EntityType: "user", EntityID: "u-2", Outcome: audit.OutcomeDenied,
		Context: map[string]any{"requested_roles": []any{"CONSELHO_FISCAL"}, "denial": "privilege_escalation"},
	}
	boom := errors.New("boom")

	err := database.WithTx(ctxWith("req-12345678", actorID), e.app, func(ctx context.Context) error {
		if err := e.rec.Record(ctx, promotion()); err != nil {
			return err
		}
		e.rec.RecordSecurity(ctx, denied)
		return boom
	})

	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	rows := e.rows(t)
	if len(rows) != 1 || rows[0].Action != "role.change_denied" || rows[0].Outcome != "denied" {
		t.Fatalf("só o evento de segurança deveria ter sobrevivido: %+v", rows)
	}
	if rows[0].RequestID == nil || *rows[0].RequestID != "req-12345678" {
		t.Errorf("request_id = %v", rows[0].RequestID)
	}
}

func TestRecordSecurityWorksWithoutAnyBusinessTransactionForAnonymousActors(t *testing.T) {
	e := newEnv(t)

	e.rec.RecordSecurity(ctxWith("req-12345678", ""), audit.Entry{
		ActorType: audit.ActorAnonymous, Action: audit.AuthLoginFailed, EntityType: "user", EntityID: "unknown",
		Outcome: audit.OutcomeFailure, Context: map[string]any{"category": "unknown_user", "email_hash": "ab12", "password": "SEGREDO"},
	})

	rows := e.rows(t)
	if len(rows) != 1 || rows[0].ActorType != "anonymous" || rows[0].ActorUserID != nil {
		t.Fatalf("linhas = %+v", rows)
	}
	if strings.Contains(rows[0].Context, "SEGREDO") {
		t.Errorf("segredo no contexto: %s", rows[0].Context)
	}
}

// AUD-04.6: se a gravação do evento falhar, vira só um incidente registrado em log.
func TestRecordSecurityFailureIsLoggedAsAnIncidentAndDoesNotPanicOrReturn(t *testing.T) {
	e := newEnv(t)
	if err := e.owner.Exec("ALTER TABLE audit_log RENAME TO audit_log_off").Error; err != nil {
		t.Fatal(err)
	}

	e.rec.RecordSecurity(ctxWith("req-12345678", ""), audit.Entry{
		ActorType: audit.ActorAnonymous, Action: audit.AuthLoginFailed, EntityType: "user", EntityID: "unknown",
		Outcome: audit.OutcomeFailure, Context: map[string]any{"password": "SEGREDO"},
	})

	out := e.logs.String()
	for _, want := range []string{`"level":"ERROR"`, "audit_security_write_failed", "auth.login_failed", "req-12345678"} {
		if !strings.Contains(out, want) {
			t.Errorf("o log deveria conter %q: %s", want, out)
		}
	}
	if strings.Contains(out, "SEGREDO") {
		t.Errorf("o log não pode conter segredos: %s", out)
	}
}

func TestRecordSecurityWithAnInvalidEntryIsLoggedAndNotWritten(t *testing.T) {
	e := newEnv(t)

	e.rec.RecordSecurity(ctxWith("req-12345678", ""), audit.Entry{
		ActorType: audit.ActorAnonymous, Action: "auth.inventada", EntityType: "user", EntityID: "x", Outcome: audit.OutcomeFailure,
	})

	if len(e.rows(t)) != 0 || !strings.Contains(e.logs.String(), "audit_security_write_failed") {
		t.Errorf("linhas = %d, log = %s", len(e.rows(t)), e.logs.String())
	}
}
