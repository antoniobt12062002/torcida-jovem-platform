//go:build integration

package infra_test

import (
	"context"
	"errors"
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
