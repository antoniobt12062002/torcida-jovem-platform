//go:build integration

package infra_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/identity/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

func newUserRepo(t *testing.T) (env, *infra.UserRepository) {
	t.Helper()
	e := newEnv(t)
	return e, infra.NewUserRepository(e.app)
}

func newDomainUser(email string) domain.User {
	return domain.User{Email: email, Name: "Ana", PasswordHash: "$argon2id$hash", Active: true}
}

func TestCreateUserStoresItAndReturnsTheIDAndTimestamps(t *testing.T) {
	_, repo := newUserRepo(t)

	u, err := repo.Create(context.Background(), newDomainUser("ana@exemplo.com"))

	if err != nil {
		t.Fatal(err)
	}
	if u.ID == "" || u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() || u.Email != "ana@exemplo.com" || !u.Active || u.MustChangePassword {
		t.Errorf("usuário = %+v", u)
	}
	got, err := repo.FindByID(context.Background(), u.ID)
	if err != nil || got.PasswordHash != "$argon2id$hash" || got.Name != "Ana" {
		t.Errorf("FindByID = %+v, %v", got, err)
	}
}

// IDN-04.2: e-mail repetido. A unicidade ignorando a caixa é do banco (índice em
// lower(email), testado na migração); o repositório só recebe e-mails normalizados.
func TestCreateUserRejectsARepeatedEmail(t *testing.T) {
	_, repo := newUserRepo(t)
	if _, err := repo.Create(context.Background(), newDomainUser("ana@exemplo.com")); err != nil {
		t.Fatal(err)
	}

	_, err := repo.Create(context.Background(), newDomainUser("ana@exemplo.com"))

	if !errors.Is(err, domain.ErrEmailTaken) {
		t.Errorf("err = %v, esperava ErrEmailTaken", err)
	}
}

func TestCreateUserRequiresANormalizedEmail(t *testing.T) {
	_, repo := newUserRepo(t)

	for _, email := range []string{"Ana@Exemplo.com", " ana@exemplo.com", "inválido", ""} {
		if _, err := repo.Create(context.Background(), newDomainUser(email)); err == nil {
			t.Errorf("%q: o repositório só grava e-mails já normalizados", email)
		}
	}
}

// IDN-04.3 e a borda de concorrência: exatamente uma criação vence.
func TestConcurrentCreatesOfTheSameEmailLeaveExactlyOneWinner(t *testing.T) {
	e, repo := newUserRepo(t)
	const attempts = 8
	errs := make([]error, attempts)

	var wg sync.WaitGroup
	for i := range attempts {
		wg.Go(func() {
			_, errs[i] = repo.Create(context.Background(), newDomainUser("corrida@exemplo.com"))
		})
	}
	wg.Wait()

	winners, taken := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			winners++
		case errors.Is(err, domain.ErrEmailTaken):
			taken++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if winners != 1 || taken != attempts-1 {
		t.Errorf("vencedores = %d, ErrEmailTaken = %d", winners, taken)
	}
	if n := count(t, e.owner, "users"); n != 1 {
		t.Errorf("usuários = %d", n)
	}
}

func TestFindByEmailIgnoresCaseAndSpaces(t *testing.T) {
	_, repo := newUserRepo(t)
	created, _ := repo.Create(context.Background(), newDomainUser("ana@exemplo.com"))

	for _, q := range []string{"ana@exemplo.com", "ANA@Exemplo.com", "  Ana@exemplo.COM  "} {
		got, err := repo.FindByEmail(context.Background(), q)
		if err != nil || got.ID != created.ID {
			t.Errorf("FindByEmail(%q) = %+v, %v", q, got, err)
		}
	}
}

func TestFindReturnsErrUserNotFound(t *testing.T) {
	_, repo := newUserRepo(t)

	if _, err := repo.FindByEmail(context.Background(), "ninguem@exemplo.com"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("FindByEmail: %v", err)
	}
	if _, err := repo.FindByEmail(context.Background(), "inválido"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("um e-mail inválido é só um usuário que não existe: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), "0f8fad5b-d9cb-469f-a165-70867728950e"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("FindByID: %v", err)
	}
	if _, err := repo.FindByID(context.Background(), "não-é-uuid"); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("um id malformado é só um usuário que não existe: %v", err)
	}
}

func TestSetPasswordAndSetActiveUpdateTheUserAndTheTimestamp(t *testing.T) {
	_, repo := newUserRepo(t)
	u, _ := repo.Create(context.Background(), newDomainUser("ana@exemplo.com"))
	time.Sleep(5 * time.Millisecond)

	if err := repo.SetPassword(context.Background(), u.ID, "$argon2id$novo", true); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetActive(context.Background(), u.ID, false); err != nil {
		t.Fatal(err)
	}

	got, _ := repo.FindByID(context.Background(), u.ID)
	if got.PasswordHash != "$argon2id$novo" || !got.MustChangePassword || got.Active || !got.UpdatedAt.After(u.UpdatedAt) {
		t.Errorf("usuário = %+v", got)
	}
	if err := repo.SetActive(context.Background(), "0f8fad5b-d9cb-469f-a165-70867728950e", true); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("SetActive de quem não existe: %v", err)
	}
	if err := repo.SetPassword(context.Background(), "0f8fad5b-d9cb-469f-a165-70867728950e", "h", false); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("SetPassword de quem não existe: %v", err)
	}
}

// A criação e a troca de papéis participam da transação do caso de uso.
func TestRepositoryJoinsTheTransactionOfTheContext(t *testing.T) {
	e, repo := newUserRepo(t)
	boom := errors.New("boom")

	err := database.WithTx(context.Background(), e.app, func(ctx context.Context) error {
		if _, err := repo.Create(ctx, newDomainUser("ana@exemplo.com")); err != nil {
			return err
		}
		return boom
	})

	if !errors.Is(err, boom) || count(t, e.owner, "users") != 0 {
		t.Errorf("err = %v, usuários = %d: a criação deveria ter sido revertida", err, count(t, e.owner, "users"))
	}
}

func TestSetRolesReplacesTheRoleSet(t *testing.T) {
	e, repo := newUserRepo(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	u, _ := repo.Create(context.Background(), newDomainUser("ana@exemplo.com"))

	if err := repo.SetRoles(context.Background(), u.ID, []domain.Role{domain.RoleAssociado, domain.RoleTesouraria}); err != nil {
		t.Fatal(err)
	}
	if err := repo.SetRoles(context.Background(), u.ID, []domain.Role{domain.RoleEventos}); err != nil {
		t.Fatal(err)
	}

	got, err := e.repo.RolesOf(context.Background(), u.ID)
	if err != nil || len(got) != 1 || got[0] != domain.RoleEventos {
		t.Errorf("papéis = %v, %v", got, err)
	}
}

func TestSetRolesRejectsAnEmptyOrUnsyncedRoleSet(t *testing.T) {
	e, repo := newUserRepo(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	u, _ := repo.Create(context.Background(), newDomainUser("ana@exemplo.com"))
	_ = repo.SetRoles(context.Background(), u.ID, []domain.Role{domain.RoleAssociado})

	if err := repo.SetRoles(context.Background(), u.ID, nil); err == nil {
		t.Error("um conjunto vazio de papéis é inválido")
	}
	if err := repo.SetRoles(context.Background(), u.ID, []domain.Role{"CHEFE"}); err == nil {
		t.Error("um papel desconhecido deveria ser recusado")
	}
	got, _ := e.repo.RolesOf(context.Background(), u.ID)
	if len(got) != 1 || got[0] != domain.RoleAssociado {
		t.Errorf("uma troca recusada não pode alterar os papéis: %v", got)
	}
}

// LockAdminSet só existe dentro de uma transação e serializa quem reduz o conjunto de administradores.
func TestLockAdminSetRequiresATransactionAndHoldsTheAdvisoryLockUntilItEnds(t *testing.T) {
	e, repo := newUserRepo(t)
	if err := repo.LockAdminSet(context.Background()); err == nil {
		t.Fatal("fora de uma transação o lock não faria sentido: deveria falhar")
	}

	tryLock := func() bool {
		var got bool
		if err := e.owner.Raw("SELECT pg_try_advisory_lock(?)", infra.AdminSetLockKey).Scan(&got).Error; err != nil {
			t.Fatal(err)
		}
		if got {
			_ = e.owner.Exec("SELECT pg_advisory_unlock(?)", infra.AdminSetLockKey).Error
		}
		return got
	}
	// pg_try_advisory_lock (de sessão) e pg_advisory_xact_lock disputam a mesma chave.
	err := database.WithTx(context.Background(), e.app, func(ctx context.Context) error {
		if err := repo.LockAdminSet(ctx); err != nil {
			return err
		}
		if tryLock() {
			t.Error("com o lock de transação preso, outra conexão não deveria conseguir a chave")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !tryLock() {
		t.Error("depois do fim da transação a chave deveria estar livre")
	}
}

func TestActiveHoldersOfCountsOnlyActiveUsersWithTheActivePermission(t *testing.T) {
	e, repo := newUserRepo(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	mk := func(email string, active bool, roles ...domain.Role) string {
		u, err := repo.Create(context.Background(), newDomainUser(email))
		if err != nil {
			t.Fatal(err)
		}
		if err := repo.SetRoles(context.Background(), u.ID, roles); err != nil {
			t.Fatal(err)
		}
		if !active {
			_ = repo.SetActive(context.Background(), u.ID, false)
		}
		return u.ID
	}
	admin := mk("admin@exemplo.com", true, domain.RoleAdminSistema)
	mk("presidente@exemplo.com", true, domain.RolePresidente, domain.RoleAssociado)
	mk("inativo@exemplo.com", false, domain.RoleAdminSistema)
	mk("diretoria@exemplo.com", true, domain.RoleDiretoria)
	mk("associado@exemplo.com", true, domain.RoleAssociado)

	const grant = "identity:admin:grant"
	if n, err := repo.ActiveHoldersOf(context.Background(), grant, ""); err != nil || n != 2 {
		t.Errorf("titulares = %d, %v (ADMIN_SISTEMA e PRESIDENTE ativos)", n, err)
	}
	if n, _ := repo.ActiveHoldersOf(context.Background(), grant, admin); n != 1 {
		t.Errorf("excluindo o admin, titulares = %d", n)
	}
	if err := e.owner.Exec("UPDATE permissions SET is_active = false WHERE name = 'identity:admin:grant'").Error; err != nil {
		t.Fatal(err)
	}
	if n, _ := repo.ActiveHoldersOf(context.Background(), grant, ""); n != 0 {
		t.Errorf("permissão inativa não conta, titulares = %d", n)
	}
}

func TestSetMustChangePasswordChangesOnlyTheFlag(t *testing.T) {
	_, repo := newUserRepo(t)
	u, _ := repo.Create(context.Background(), newDomainUser("ana@exemplo.com"))

	if err := repo.SetMustChangePassword(context.Background(), u.ID, true); err != nil {
		t.Fatal(err)
	}

	got, _ := repo.FindByID(context.Background(), u.ID)
	if !got.MustChangePassword || got.PasswordHash != u.PasswordHash || !got.Active {
		t.Errorf("usuário = %+v", got)
	}
	if err := repo.SetMustChangePassword(context.Background(), "0f8fad5b-d9cb-469f-a165-70867728950e", true); !errors.Is(err, domain.ErrUserNotFound) {
		t.Errorf("err = %v", err)
	}
}

func mkListed(t *testing.T, e env, repo *infra.UserRepository, email string, roles ...domain.Role) domain.User {
	t.Helper()
	u, err := repo.Create(context.Background(), newDomainUser(email))
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) == 0 {
		roles = []domain.Role{domain.RoleAssociado}
	}
	if err := repo.SetRoles(context.Background(), u.ID, roles); err != nil {
		t.Fatal(err)
	}
	return u
}

// IDN-04.8: cursor por (created_at, id), do mais novo para o mais antigo.
func TestListPagesByCursorNewestFirstWithoutRepetitionOrGaps(t *testing.T) {
	e, repo := newUserRepo(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	var ids []string
	for i := range 5 {
		ids = append(ids, mkListed(t, e, repo, "u"+string(rune('a'+i))+"@exemplo.com").ID)
	}

	var got []string
	var cursor *domain.ListCursor
	for range 5 {
		page, err := repo.List(context.Background(), domain.ListFilter{}, cursor, 2)
		if err != nil {
			t.Fatal(err)
		}
		for _, it := range page {
			got = append(got, it.ID)
		}
		if len(page) < 2 {
			break
		}
		last := page[len(page)-1]
		cursor = &domain.ListCursor{CreatedAt: last.CreatedAt, ID: last.ID}
	}

	want := slices.Clone(ids)
	slices.Reverse(want)
	if !slices.Equal(got, want) {
		t.Errorf("ordem = %v, esperado %v (mais novo primeiro)", got, want)
	}
}

func TestListFiltersByActiveAndRoleAndCarriesRolesAndTheMembershipSummary(t *testing.T) {
	e, repo := newUserRepo(t)
	if _, err := e.repo.Sync(context.Background(), matrix(t)); err != nil {
		t.Fatal(err)
	}
	members := infra.NewAdminMembershipRepository(e.app)
	tes := mkListed(t, e, repo, "tes@exemplo.com", domain.RoleAssociado, domain.RoleTesouraria)
	m, _ := domain.NewAdminMembership(tes.ID, "eleito tesoureiro na assembleia", nil, time.Now().UTC())
	if _, err := members.Grant(context.Background(), m); err != nil {
		t.Fatal(err)
	}
	off := mkListed(t, e, repo, "off@exemplo.com")
	_ = repo.SetActive(context.Background(), off.ID, false)
	mkListed(t, e, repo, "plain@exemplo.com")

	yes, no := true, false
	role := domain.RoleTesouraria
	byRole, err := repo.List(context.Background(), domain.ListFilter{Role: &role}, nil, 10)
	if err != nil || len(byRole) != 1 || byRole[0].ID != tes.ID {
		t.Fatalf("por papel = %+v, %v", byRole, err)
	}
	if got := byRole[0]; got.AdminMembership == nil || got.AdminMembership.Reason != "eleito tesoureiro na assembleia" ||
		!slices.Contains(got.Roles, domain.RoleTesouraria) || !slices.Contains(got.Roles, domain.RoleAssociado) || got.Email != "tes@exemplo.com" {
		t.Errorf("item = %+v", got)
	}
	if inactive, _ := repo.List(context.Background(), domain.ListFilter{Active: &no}, nil, 10); len(inactive) != 1 || inactive[0].ID != off.ID || inactive[0].Active {
		t.Errorf("inativos = %+v", inactive)
	}
	if active, _ := repo.List(context.Background(), domain.ListFilter{Active: &yes}, nil, 10); len(active) != 2 {
		t.Errorf("ativos = %d", len(active))
	}
	all, _ := repo.List(context.Background(), domain.ListFilter{}, nil, 10)
	if len(all) != 3 {
		t.Errorf("todos = %d", len(all))
	}
	for _, it := range all {
		if it.ID != tes.ID && it.AdminMembership != nil {
			t.Errorf("só quem tem vínculo ativo traz o resumo: %+v", it)
		}
	}
}

func TestListWithNoUsersIsEmpty(t *testing.T) {
	_, repo := newUserRepo(t)

	if page, err := repo.List(context.Background(), domain.ListFilter{}, nil, 10); err != nil || len(page) != 0 {
		t.Errorf("page = %v, err = %v", page, err)
	}
}
