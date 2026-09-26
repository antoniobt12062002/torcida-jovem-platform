//go:build integration

package infra_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
)

type membershipEnv struct {
	env
	users *infra.UserRepository
	repo  *infra.AdminMembershipRepository
}

func newMembershipEnv(t *testing.T) membershipEnv {
	t.Helper()
	e := newEnv(t)
	return membershipEnv{env: e, users: infra.NewUserRepository(e.app), repo: infra.NewAdminMembershipRepository(e.app)}
}

func (m membershipEnv) user(t *testing.T, email string) domain.User {
	t.Helper()
	u, err := m.users.Create(context.Background(), newDomainUser(email))
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func grantTo(t *testing.T, userID string, grantedBy *string) domain.AdminMembership {
	t.Helper()
	m, err := domain.NewAdminMembership(userID, "eleito tesoureiro na assembleia", grantedBy, time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestGrantStoresTheMembershipAndFindsItAsActive(t *testing.T) {
	e := newMembershipEnv(t)
	admin, target := e.user(t, "admin@exemplo.com"), e.user(t, "ana@exemplo.com")

	saved, err := e.repo.Grant(context.Background(), grantTo(t, target.ID, &admin.ID))

	if err != nil || saved.ID == "" || !saved.Active() {
		t.Fatalf("Grant = %+v, %v", saved, err)
	}
	active, ok, err := e.repo.Active(context.Background(), target.ID)
	if err != nil || !ok || active.ID != saved.ID || active.GrantedBy == nil || *active.GrantedBy != admin.ID || active.Reason != "eleito tesoureiro na assembleia" {
		t.Errorf("Active = %+v, %v, %v", active, ok, err)
	}
	if _, ok, _ := e.repo.Active(context.Background(), admin.ID); ok {
		t.Error("quem não tem vínculo não tem vínculo ativo")
	}
}

func TestBootstrapMembershipHasNoGrantorInTheDatabase(t *testing.T) {
	e := newMembershipEnv(t)
	u := e.user(t, "admin@exemplo.com")
	m, _ := domain.NewAdminMembership(u.ID, domain.BootstrapReason, nil, time.Now().UTC())

	saved, err := e.repo.Grant(context.Background(), m)

	if err != nil || saved.GrantedBy != nil || saved.Reason != "bootstrap-admin" {
		t.Errorf("Grant = %+v, %v", saved, err)
	}
}

// IDN-06.3: dois vínculos ativos ao mesmo tempo, exatamente um vence.
func TestConcurrentGrantsToTheSameUserLeaveExactlyOneActiveMembership(t *testing.T) {
	e := newMembershipEnv(t)
	target := e.user(t, "ana@exemplo.com")
	const attempts = 6
	errs := make([]error, attempts)

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() { _, errs[i] = e.repo.Grant(context.Background(), grantTo(t, target.ID, nil)) })
	}
	wg.Wait()

	winners, already := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, domain.ErrAlreadyAdmin):
			already++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if winners != 1 || already != attempts-1 {
		t.Errorf("vencedores = %d, ErrAlreadyAdmin = %d", winners, already)
	}
}

// IDN-06.5: encerrar preserva a linha e o histórico continua consultável.
func TestRevokePreservesTheRowAndTheHistoryStaysQueryable(t *testing.T) {
	e := newMembershipEnv(t)
	admin, target := e.user(t, "admin@exemplo.com"), e.user(t, "ana@exemplo.com")
	first, _ := e.repo.Grant(context.Background(), grantTo(t, target.ID, &admin.ID))

	closed, err := e.repo.Revoke(context.Background(), target.ID, admin.ID, "fim do mandato", time.Now().UTC())
	if err != nil {
		t.Fatal(err)
	}
	second, err := e.repo.Grant(context.Background(), grantTo(t, target.ID, &admin.ID))
	if err != nil {
		t.Fatalf("depois de encerrado, um novo vínculo deveria ser aceito: %v", err)
	}

	if closed.ID != first.ID || closed.Active() || closed.RevokedBy == nil || *closed.RevokedBy != admin.ID ||
		closed.RevokeReason == nil || *closed.RevokeReason != "fim do mandato" || closed.RevokedAt == nil {
		t.Errorf("encerrado = %+v", closed)
	}
	history, err := e.repo.History(context.Background(), target.ID)
	if err != nil || len(history) != 2 || history[0].ID != second.ID || history[1].ID != first.ID {
		t.Fatalf("histórico = %+v, %v (do mais novo para o mais antigo)", history, err)
	}
	if history[1].Active() || history[1].Reason != "eleito tesoureiro na assembleia" || !history[0].Active() {
		t.Errorf("o histórico deveria manter o vínculo antigo intacto: %+v", history)
	}
	if n := count(t, e.owner, "admin_memberships"); n != 2 {
		t.Errorf("linhas = %d, nenhuma pode ser apagada", n)
	}
}

func TestRevokeWithoutAnActiveMembershipReturnsErrNotAdmin(t *testing.T) {
	e := newMembershipEnv(t)
	admin, target := e.user(t, "admin@exemplo.com"), e.user(t, "ana@exemplo.com")

	if _, err := e.repo.Revoke(context.Background(), target.ID, admin.ID, "fim do mandato", time.Now().UTC()); !errors.Is(err, domain.ErrNotAdmin) {
		t.Errorf("err = %v", err)
	}
	if _, err := e.repo.Revoke(context.Background(), "não-é-uuid", admin.ID, "fim do mandato", time.Now().UTC()); !errors.Is(err, domain.ErrNotAdmin) {
		t.Errorf("id malformado: %v", err)
	}
}

func TestRevokeRequiresAReason(t *testing.T) {
	e := newMembershipEnv(t)
	admin, target := e.user(t, "admin@exemplo.com"), e.user(t, "ana@exemplo.com")
	_, _ = e.repo.Grant(context.Background(), grantTo(t, target.ID, &admin.ID))

	if _, err := e.repo.Revoke(context.Background(), target.ID, admin.ID, "   ", time.Now().UTC()); !errors.Is(err, domain.ErrReasonRequired) {
		t.Errorf("err = %v", err)
	}
	if _, ok, _ := e.repo.Active(context.Background(), target.ID); !ok {
		t.Error("um encerramento recusado não pode fechar o vínculo")
	}
}

func TestHistoryOfAUserWithoutMembershipsIsEmpty(t *testing.T) {
	e := newMembershipEnv(t)
	u := e.user(t, "ana@exemplo.com")

	if h, err := e.repo.History(context.Background(), u.ID); err != nil || len(h) != 0 {
		t.Errorf("histórico = %v, %v", h, err)
	}
	if h, err := e.repo.History(context.Background(), "não-é-uuid"); err != nil || len(h) != 0 {
		t.Errorf("id malformado: %v, %v", h, err)
	}
}

// IDN-06.5: o repositório não expõe operação de exclusão.
func TestRepositoryExposesNoDeleteOperation(t *testing.T) {
	typ := reflect.TypeOf(&infra.AdminMembershipRepository{})
	for i := range typ.NumMethod() {
		name := strings.ToLower(typ.Method(i).Name)
		if strings.Contains(name, "delete") || strings.Contains(name, "remove") || strings.Contains(name, "purge") {
			t.Errorf("o método %s não deveria existir: o histórico é preservado", typ.Method(i).Name)
		}
	}
}

// Duas retiradas simultâneas: uma encerra, a outra vê que não há vínculo ativo.
func TestConcurrentRevokesCloseTheMembershipOnlyOnce(t *testing.T) {
	e := newMembershipEnv(t)
	admin, target := e.user(t, "admin@exemplo.com"), e.user(t, "ana@exemplo.com")
	_, _ = e.repo.Grant(context.Background(), grantTo(t, target.ID, &admin.ID))
	const attempts = 6
	errs := make([]error, attempts)

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() {
			_, errs[i] = e.repo.Revoke(context.Background(), target.ID, admin.ID, "fim do mandato", time.Now().UTC())
		})
	}
	wg.Wait()

	winners := 0
	for _, err := range errs {
		switch {
		case err == nil:
			winners++
		case !errors.Is(err, domain.ErrNotAdmin):
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if winners != 1 {
		t.Errorf("exatamente uma retirada deveria vencer, vencedoras = %d", winners)
	}
}
