//go:build integration

package infra_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

func newExistenceCheckerEnv(t *testing.T) (*gorm.DB, *infra.ContaRepository, *infra.LancamentoRepository, *infra.LancamentoExistenceChecker) {
	t.Helper()
	db := testutil.NewTestDB(t)
	return db, infra.NewContaRepository(db), infra.NewLancamentoRepository(db), infra.NewLancamentoExistenceChecker(db)
}

func newUserForChecker(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var id string
	err := db.Raw(`INSERT INTO users (email, name, password_hash) VALUES (?, 'Fulano', 'hash') RETURNING id::text`,
		uuid.New().String()+"@exemplo.com").Scan(&id).Error
	if err != nil || id == "" {
		t.Fatalf("criar usuário: id=%q err=%v", id, err)
	}
	return id
}

// PC-02 AC1 (consumida por 01/T4): conta nunca usada — TemLancamento
// retorna falso.
func TestLancamentoExistenceCheckerReturnsFalseForAnUnusedConta(t *testing.T) {
	_, contas, _, checker := newExistenceCheckerEnv(t)
	conta, err := contas.Criar(context.Background(), domain.Conta{Tipo: domain.TipoReceita, Nome: "Produtos"})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	tem, err := checker.TemLancamento(context.Background(), conta.ID)
	if err != nil {
		t.Fatalf("TemLancamento: %v", err)
	}
	if tem {
		t.Error("esperava false para conta nunca usada")
	}
}

// PC-02 AC2 (consumida por 01/T4): conta já usada, mesmo com o único
// lançamento CANCELADA — TemLancamento retorna verdadeiro.
func TestLancamentoExistenceCheckerReturnsTrueEvenWhenTheOnlyLancamentoIsCancelled(t *testing.T) {
	db, contas, lancamentos, checker := newExistenceCheckerEnv(t)
	conta, err := contas.Criar(context.Background(), domain.Conta{Tipo: domain.TipoDespesa, Nome: "Material"})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}
	userID := newUserForChecker(t, db)
	lancamento, err := lancamentos.Criar(context.Background(), domain.Lancamento{
		Tipo: domain.TipoDespesa, ContaID: conta.ID, ValorBrutoCents: 100, ValorLiquidoCents: 100,
		FormaPagamento: domain.FormaPix, CriadoPor: userID,
	})
	if err != nil {
		t.Fatalf("criar lançamento: %v", err)
	}
	if err := db.Exec("UPDATE lancamentos SET status = 'CANCELADA' WHERE id = ?::uuid", lancamento.ID).Error; err != nil {
		t.Fatalf("fixture: marcar CANCELADA: %v", err)
	}

	tem, err := checker.TemLancamento(context.Background(), conta.ID)
	if err != nil {
		t.Fatalf("TemLancamento: %v", err)
	}
	if !tem {
		t.Error("esperava true mesmo com o único lançamento CANCELADA")
	}
}

// Nunca escreve em contas_contabeis: uma segunda leitura da mesma conta
// continua exatamente igual.
func TestLancamentoExistenceCheckerNeverWritesToContasContabeis(t *testing.T) {
	_, contas, _, checker := newExistenceCheckerEnv(t)
	conta, err := contas.Criar(context.Background(), domain.Conta{Tipo: domain.TipoReceita, Nome: "Produtos"})
	if err != nil {
		t.Fatalf("criar conta: %v", err)
	}

	if _, err := checker.TemLancamento(context.Background(), conta.ID); err != nil {
		t.Fatalf("TemLancamento: %v", err)
	}

	depois, err := contas.Buscar(context.Background(), conta.ID)
	if err != nil {
		t.Fatalf("buscar conta: %v", err)
	}
	if depois.Nome != conta.Nome || depois.Tipo != conta.Tipo || depois.Ativo != conta.Ativo {
		t.Errorf("conta foi alterada: antes=%+v depois=%+v", conta, depois)
	}
}
