//go:build integration

package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/authz"
)

func (e env) criarConta() *app.CriarConta {
	return &app.CriarConta{Authz: simpleAuthz{}, Contas: e.contas, Audit: e.rec, Tx: e.tx}
}

// PC-01 AC1/AC2: cria conta raiz e subconta com o mesmo tipo do pai.
func TestCriarContaPersistsRootAndSubcontaWithTheSameTipo(t *testing.T) {
	e := newEnv(t)

	raiz, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if err != nil {
		t.Fatalf("criar conta raiz: %v", err)
	}
	if raiz.ID == "" || raiz.Tipo != domain.TipoReceita || raiz.Nome != "Produtos" || !raiz.Ativo || raiz.ParentID != nil {
		t.Errorf("conta raiz = %+v", raiz)
	}

	sub, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Camisetas", ParentID: &raiz.ID,
	})
	if err != nil {
		t.Fatalf("criar subconta: %v", err)
	}
	if sub.ParentID == nil || *sub.ParentID != raiz.ID {
		t.Errorf("subconta.ParentID = %v, esperado %q", sub.ParentID, raiz.ID)
	}
}

// PC-01 AC2: subconta com tipo diferente do pai é recusada.
func TestCriarContaRejectsASubcontaWithADifferentTipoThanTheParent(t *testing.T) {
	e := newEnv(t)
	raiz, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if err != nil {
		t.Fatalf("criar conta raiz: %v", err)
	}

	_, err = e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoDespesa, Nome: "Frete", ParentID: &raiz.ID,
	})
	if !errors.Is(err, domain.ErrContaTipoIncompativel) {
		t.Errorf("err = %v, esperado ErrContaTipoIncompativel", err)
	}
}

// PC-01 AC2: parent_id inexistente é recusado.
func TestCriarContaRejectsANonexistentParent(t *testing.T) {
	e := newEnv(t)
	bogus := "00000000-0000-0000-0000-000000000000"

	_, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoReceita, Nome: "x", ParentID: &bogus,
	})
	if !errors.Is(err, domain.ErrContaNaoEncontrada) {
		t.Errorf("err = %v, esperado ErrContaNaoEncontrada", err)
	}
}

// PC-01 AC3: sem a permissão, nada é escrito.
func TestCriarContaRequiresThePermissionAndWritesNothingWithoutIt(t *testing.T) {
	e := newEnv(t)

	_, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(), Tipo: domain.TipoReceita, Nome: "Produtos",
	})
	if !errors.Is(err, authz.ErrForbidden) {
		t.Errorf("err = %v, esperado authz.ErrForbidden", err)
	}
	var n int64
	if err := e.db.Raw("SELECT count(*) FROM contas_contabeis").Scan(&n).Error; err != nil || n != 0 {
		t.Errorf("nenhuma conta deveria ter sido gravada: n=%d err=%v", n, err)
	}
}

// PC-01 AC4: a criação audita id, tipo e nome.
func TestCriarContaAuditsIDTipoAndNome(t *testing.T) {
	e := newEnv(t)

	conta, err := e.criarConta().Execute(context.Background(), app.CriarContaInput{
		Actor: actor(app.PermContaCreate), Tipo: domain.TipoDespesa, Nome: "Material",
	})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	var action, entityID, after string
	err = e.db.Raw(`SELECT action, entity_id, after::text FROM audit_log WHERE action = 'conta.create' AND entity_id = ?`, conta.ID).
		Row().Scan(&action, &entityID, &after)
	if err != nil {
		t.Fatalf("ler auditoria: %v", err)
	}
	if entityID != conta.ID || after == "" {
		t.Errorf("auditoria não confere: entity_id=%q after=%q", entityID, after)
	}
}
