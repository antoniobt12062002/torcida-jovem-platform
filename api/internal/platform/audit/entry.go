// Package audit records the immutable audit trail (ADR-004).
package audit

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

// Redacted replaces the value of a sensitive key before an entry is stored.
const Redacted = "[redacted]"

type ActorType string

const (
	ActorUser      ActorType = "user"
	ActorSystem    ActorType = "system"
	ActorAnonymous ActorType = "anonymous"
)

type Outcome string

const (
	OutcomeSuccess Outcome = "success"
	OutcomeDenied  Outcome = "denied"
	OutcomeFailure Outcome = "failure"
)

// Action names an audited event, as "domain.verb".
type Action string

// Actions of the foundation. Each module registers its own with Register.
const (
	UserCreate         Action = "user.create"
	UserBootstrap      Action = "user.bootstrap"
	UserDeactivate     Action = "user.deactivate"
	UserPasswordChange Action = "user.password_change"
	UserRolesSet       Action = "user.roles_set"
	AdminPromote       Action = "admin.promote"
	AdminRevoke        Action = "admin.revoke"
	RBACSync           Action = "rbac.sync"
	RoleChangeDenied   Action = "role.change_denied"
	AuthLogin          Action = "auth.login"
	AuthLoginFailed    Action = "auth.login_failed"
	AuthLoginBlocked   Action = "auth.login_blocked"
	AuthLogout         Action = "auth.logout"
	AuthzDenied        Action = "authz.denied"

	AuthPasswordResetRequested Action = "auth.password_reset_requested"
	AuthPasswordResetCompleted Action = "auth.password_reset_completed"
	AuthPasswordResetFailed    Action = "auth.password_reset_failed"
	UserPasswordReset          Action = "user.password_reset"
	UserReactivate             Action = "user.reactivate"
)

var (
	actionFormat = regexp.MustCompile(`^[a-z_]+\.[a-z_]+$`)
	uuidFormat   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

	catalogMu sync.RWMutex
	catalog   = map[Action]struct{}{
		UserCreate: {}, UserBootstrap: {}, UserDeactivate: {}, UserPasswordChange: {}, UserRolesSet: {},
		AdminPromote: {}, AdminRevoke: {}, RBACSync: {}, RoleChangeDenied: {},
		AuthLogin: {}, AuthLoginFailed: {}, AuthLoginBlocked: {}, AuthLogout: {}, AuthzDenied: {},
		AuthPasswordResetRequested: {}, AuthPasswordResetCompleted: {}, AuthPasswordResetFailed: {}, UserPasswordReset: {}, UserReactivate: {},
	}

	sensitiveKeys = map[string]struct{}{
		"password": {}, "password_hash": {}, "temporary_password": {}, "token": {}, "session_token": {},
		"csrf_token": {}, "cookie": {}, "authorization": {},
	}
)

// Register adds actions to the catalog. It panics on a name that is not
// "domain.verb", since that is a programming error caught at startup.
func Register(actions ...Action) {
	catalogMu.Lock()
	defer catalogMu.Unlock()
	for _, a := range actions {
		if !actionFormat.MatchString(string(a)) {
			panic(fmt.Sprintf("audit: ação %q fora do formato dominio.verbo", a))
		}
		catalog[a] = struct{}{}
	}
}

// Known reports whether the action is in the catalog.
func (a Action) Known() bool {
	catalogMu.RLock()
	defer catalogMu.RUnlock()
	_, ok := catalog[a]
	return ok
}

// requiresReason: cancellations and extraordinary adjustments (ADR-004), and
// the grant and withdrawal of administrative access (IDN-06).
func (a Action) requiresReason() bool {
	s := string(a)
	return strings.HasSuffix(s, ".cancel") || strings.HasSuffix(s, ".adjust") || a == AdminPromote || a == AdminRevoke
}

// Entry is one audit record. EntityType and EntityID identify the target.
type Entry struct {
	ActorType  ActorType
	ActorID    string // the user id; only with ActorUser
	Action     Action
	EntityType string
	EntityID   string
	Before     map[string]any
	After      map[string]any
	Outcome    Outcome
	Reason     string
	Context    map[string]any
}

// ErrInvalidEntry wraps every validation failure.
var ErrInvalidEntry = errors.New("audit: entrada inválida")

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalidEntry, fmt.Sprintf(format, args...))
}

// Validate checks the entry before it is written.
func (e Entry) Validate() error {
	switch e.ActorType {
	case ActorUser:
		if !uuidFormat.MatchString(e.ActorID) {
			return invalid("ator do tipo user exige o id do usuário (UUID)")
		}
	case ActorSystem, ActorAnonymous:
		if e.ActorID != "" {
			return invalid("ator do tipo %s não tem id de usuário", e.ActorType)
		}
	default:
		return invalid("tipo de ator %q desconhecido", e.ActorType)
	}
	switch e.Outcome {
	case OutcomeSuccess, OutcomeDenied, OutcomeFailure:
	default:
		return invalid("resultado %q desconhecido", e.Outcome)
	}
	if !e.Action.Known() {
		return invalid("ação %q fora do catálogo", e.Action)
	}
	if e.EntityType == "" || e.EntityID == "" {
		return invalid("o alvo (tipo e id) é obrigatório")
	}
	if e.Action.requiresReason() && strings.TrimSpace(e.Reason) == "" {
		return invalid("a ação %s exige motivo", e.Action)
	}
	return nil
}

// Redacted returns a copy whose before, after and context have the value of
// every sensitive key replaced by Redacted, at any depth. The receiver is not
// changed.
func (e Entry) Redacted() Entry {
	e.Before = redactMap(e.Before)
	e.After = redactMap(e.After)
	e.Context = redactMap(e.Context)
	return e
}

func redactMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		if _, sensitive := sensitiveKeys[strings.ToLower(k)]; sensitive {
			out[k] = Redacted
			continue
		}
		out[k] = redactValue(v)
	}
	return out
}

func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		return redactMap(t)
	case []map[string]any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = redactMap(item)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, item := range t {
			out[i] = redactValue(item)
		}
		return out
	default:
		return v
	}
}
