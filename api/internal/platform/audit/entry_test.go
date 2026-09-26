package audit

import (
	"encoding/json"
	"strings"
	"testing"
)

const someUser = "0f8fad5b-d9cb-469f-a165-70867728950e"

func validEntry() Entry {
	return Entry{
		ActorType:  ActorUser,
		ActorID:    someUser,
		Action:     UserCreate,
		EntityType: "user",
		EntityID:   "u-1",
		Outcome:    OutcomeSuccess,
	}
}

func TestValidEntryPasses(t *testing.T) {
	if err := validEntry().Validate(); err != nil {
		t.Errorf("entrada válida recusada: %v", err)
	}
}

// AUD-04.7: só entram ações do catálogo declarado em código.
func TestValidateRejectsActionsOutsideTheCatalog(t *testing.T) {
	e := validEntry()
	e.Action = "user.inventada"

	if err := e.Validate(); err == nil || !strings.Contains(err.Error(), "catálogo") {
		t.Errorf("esperava recusa por ação fora do catálogo, veio %v", err)
	}
}

func TestCatalogHasTheIdentityAndSecurityActionsOfTheSpec(t *testing.T) {
	for _, a := range []Action{
		UserCreate, UserBootstrap, UserDeactivate, UserPasswordChange, UserRolesSet,
		AdminPromote, AdminRevoke, RBACSync, RoleChangeDenied,
		AuthLogin, AuthLoginFailed, AuthLoginBlocked, AuthLogout,
	} {
		if !a.Known() {
			t.Errorf("a ação %q deveria estar no catálogo", a)
		}
	}
}

func TestRegisterAddsModuleActionsAndRejectsMalformedNames(t *testing.T) {
	Register("teste.criar")
	e := validEntry()
	e.Action = "teste.criar"
	if err := e.Validate(); err != nil {
		t.Errorf("ação registrada foi recusada: %v", err)
	}

	defer func() {
		if recover() == nil {
			t.Error("Register deveria recusar um nome fora do formato dominio.verbo")
		}
	}()
	Register("SemPonto")
}

// AUD-01.4: cancelamento e ajuste extraordinário exigem motivo.
func TestValidateRequiresAReasonForCancellationAndAdjustment(t *testing.T) {
	Register("teste.cancel", "teste.adjust")
	for _, action := range []Action{"teste.cancel", "teste.adjust", AdminPromote, AdminRevoke} {
		e := validEntry()
		e.Action = action
		if err := e.Validate(); err == nil || !strings.Contains(err.Error(), "motivo") {
			t.Errorf("%s sem motivo deveria ser recusada, veio %v", action, err)
		}
		e.Reason = "   "
		if err := e.Validate(); err == nil {
			t.Errorf("%s com motivo em branco deveria ser recusada", action)
		}
		e.Reason = "aprovado em assembleia"
		if err := e.Validate(); err != nil {
			t.Errorf("%s com motivo foi recusada: %v", action, err)
		}
	}
}

func TestValidateChecksActorConsistency(t *testing.T) {
	cases := []struct {
		name string
		edit func(*Entry)
	}{
		{"usuário sem id", func(e *Entry) { e.ActorID = "" }},
		{"id que não é UUID", func(e *Entry) { e.ActorID = "123" }},
		{"anônimo com id", func(e *Entry) { e.ActorType = ActorAnonymous }},
		{"sistema com id", func(e *Entry) { e.ActorType = ActorSystem }},
		{"tipo de ator inválido", func(e *Entry) { e.ActorType = "robot"; e.ActorID = "" }},
		{"resultado inválido", func(e *Entry) { e.Outcome = "maybe" }},
		{"alvo sem tipo", func(e *Entry) { e.EntityType = "" }},
		{"alvo sem id", func(e *Entry) { e.EntityID = "" }},
	}
	for _, tc := range cases {
		e := validEntry()
		tc.edit(&e)
		if err := e.Validate(); err == nil {
			t.Errorf("%s: deveria ser recusada", tc.name)
		}
	}
	for _, at := range []ActorType{ActorSystem, ActorAnonymous} {
		e := validEntry()
		e.ActorType, e.ActorID = at, ""
		if err := e.Validate(); err != nil {
			t.Errorf("ator %s sem id foi recusado: %v", at, err)
		}
	}
}

// AUD-01.5: chaves sensíveis viram [redacted] em before, after e context.
func TestRedactedHidesSensitiveKeysEverywhereWithoutTouchingTheOriginal(t *testing.T) {
	e := validEntry()
	e.Before = map[string]any{"Password": "SEGREDO-1", "name": "Ana"}
	e.After = map[string]any{
		"password_hash": "SEGREDO-2",
		"nested":        map[string]any{"session_token": "SEGREDO-3", "keep": 1},
		"list":          []any{map[string]any{"csrf_token": "SEGREDO-4"}},
	}
	e.Context = map[string]any{"token": "SEGREDO-5", "cookie": "SEGREDO-6", "authorization": "SEGREDO-7", "roles": []string{"a"}}

	got := e.Redacted()

	raw, _ := json.Marshal([]any{got.Before, got.After, got.Context})
	if strings.Contains(string(raw), "SEGREDO") {
		t.Errorf("um segredo vazou: %s", raw)
	}
	if got.Before["name"] != "Ana" || got.After["nested"].(map[string]any)["keep"] != 1 {
		t.Errorf("valores não sensíveis devem ser preservados: %s", raw)
	}
	if got.Before["Password"] != Redacted || got.Context["token"] != Redacted {
		t.Errorf("before = %v, context = %v", got.Before, got.Context)
	}
	if e.Before["Password"] != "SEGREDO-1" {
		t.Error("a entrada original não pode ser alterada")
	}
}

func TestValidateAcceptsEveryOutcome(t *testing.T) {
	for _, o := range []Outcome{OutcomeSuccess, OutcomeDenied, OutcomeFailure} {
		e := validEntry()
		e.Outcome = o
		if err := e.Validate(); err != nil {
			t.Errorf("resultado %s recusado: %v", o, err)
		}
	}
}
