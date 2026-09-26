package domain

import (
	"errors"
	"testing"
	"time"
)

var now = time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)

const grantor = "0f8fad5b-d9cb-469f-a165-70867728950e"

// IDN-06.2: o motivo é obrigatório, com ao menos 10 caracteres depois do trim.
func TestNewAdminMembershipRequiresAReasonOfAtLeastTenCharacters(t *testing.T) {
	for _, reason := range []string{"", "          ", "  curto   ", "123456789"} {
		_, err := NewAdminMembership("u-1", reason, new(grantor), now)
		if !errors.Is(err, ErrReasonRequired) {
			t.Errorf("motivo %q: err = %v, esperava ErrReasonRequired", reason, err)
		}
	}
	m, err := NewAdminMembership("u-1", "  eleito tesoureiro  ", new(grantor), now)
	if err != nil {
		t.Fatalf("motivo suficiente foi recusado: %v", err)
	}
	if m.Reason != "eleito tesoureiro" || m.UserID != "u-1" || m.GrantedBy == nil || *m.GrantedBy != grantor || !m.GrantedAt.Equal(now) || !m.Active() {
		t.Errorf("vínculo = %+v", m)
	}
}

func TestBootstrapMembershipHasNoGrantor(t *testing.T) {
	m, err := NewAdminMembership("u-1", BootstrapReason, nil, now)

	if err != nil || m.GrantedBy != nil || !m.Active() {
		t.Errorf("vínculo = %+v, err = %v", m, err)
	}
	if BootstrapReason != "bootstrap" {
		t.Errorf("BootstrapReason = %q", BootstrapReason)
	}
}

func TestBootstrapReasonIsAcceptedEvenThoughItIsShort(t *testing.T) {
	if _, err := NewAdminMembership("u-1", BootstrapReason, nil, now); err != nil {
		t.Errorf("o motivo de bootstrap é a exceção ao mínimo: %v", err)
	}
	if _, err := NewAdminMembership("u-1", BootstrapReason, new(grantor), now); !errors.Is(err, ErrReasonRequired) {
		t.Errorf("com quem concedeu, o motivo curto continua inválido: %v", err)
	}
}

// IDN-06.4: encerrar registra quem, quando e por quê, e não apaga.
func TestRevokeRecordsWhoWhenAndWhyAndKeepsTheHistory(t *testing.T) {
	m, _ := NewAdminMembership("u-1", "eleito tesoureiro", new(grantor), now)
	later := now.Add(48 * time.Hour)

	closed, err := m.Revoke("0f8fad5b-d9cb-469f-a165-70867728950f", "fim do mandato", later)

	if err != nil {
		t.Fatal(err)
	}
	if closed.Active() || closed.RevokedBy == nil || closed.RevokeReason == nil || *closed.RevokeReason != "fim do mandato" ||
		closed.RevokedAt == nil || !closed.RevokedAt.Equal(later) {
		t.Errorf("encerrado = %+v", closed)
	}
	if closed.Reason != m.Reason || closed.GrantedBy == nil || !closed.GrantedAt.Equal(now) {
		t.Error("o encerramento não pode alterar o que já estava registrado")
	}
	if !m.Active() {
		t.Error("o valor original não pode ser alterado")
	}
}

func TestRevokeRequiresAReasonAndAnActiveMembership(t *testing.T) {
	m, _ := NewAdminMembership("u-1", "eleito tesoureiro", new(grantor), now)

	if _, err := m.Revoke(grantor, "  ", now); !errors.Is(err, ErrReasonRequired) {
		t.Errorf("sem motivo: %v", err)
	}
	if _, err := m.Revoke("", "fim do mandato", now); err == nil {
		t.Error("é preciso dizer quem encerra")
	}
	closed, _ := m.Revoke(grantor, "fim do mandato", now)
	if _, err := closed.Revoke(grantor, "de novo", now); !errors.Is(err, ErrNotAdmin) {
		t.Errorf("encerrar duas vezes: %v", err)
	}
}

// IDN-06.3: um vínculo ativo por usuário.
func TestCheckCanGrantRefusesWhenThereIsAnActiveMembership(t *testing.T) {
	active, _ := NewAdminMembership("u-1", "eleito tesoureiro", new(grantor), now)
	closed, _ := active.Revoke(grantor, "fim do mandato", now)

	if err := CheckCanGrant([]AdminMembership{closed}); err != nil {
		t.Errorf("só vínculos encerrados não impedem: %v", err)
	}
	if err := CheckCanGrant(nil); err != nil {
		t.Errorf("sem histórico: %v", err)
	}
	if err := CheckCanGrant([]AdminMembership{closed, active}); !errors.Is(err, ErrAlreadyAdmin) {
		t.Errorf("com vínculo ativo: %v", err)
	}
}

// IDN-06.9: papel diferente de ASSOCIADO exige vínculo ativo.
func TestNonAssociadoRolesRequireAnActiveMembership(t *testing.T) {
	if err := CheckRolesHaveMembership([]Role{RoleAssociado}, false); err != nil {
		t.Errorf("ASSOCIADO não exige vínculo: %v", err)
	}
	for _, roles := range [][]Role{{RoleTesouraria}, {RoleAssociado, RoleEventos}, {RolePresidente}} {
		if err := CheckRolesHaveMembership(roles, false); !errors.Is(err, ErrAdminMembershipRequired) {
			t.Errorf("%v sem vínculo: %v", roles, err)
		}
		if err := CheckRolesHaveMembership(roles, true); err != nil {
			t.Errorf("%v com vínculo: %v", roles, err)
		}
	}
	if err := CheckRolesHaveMembership(nil, false); err != nil {
		t.Errorf("sem papéis não há o que exigir: %v", err)
	}
}

func TestErrorCodesMatchTheAPI(t *testing.T) {
	for err, code := range map[error]string{
		ErrReasonRequired: "reason_required", ErrAlreadyAdmin: "already_admin", ErrNotAdmin: "not_admin",
		ErrAdminMembershipRequired: "admin_membership_required",
	} {
		if err.Error() != code {
			t.Errorf("%v deveria ter o código %q", err, code)
		}
	}
}
