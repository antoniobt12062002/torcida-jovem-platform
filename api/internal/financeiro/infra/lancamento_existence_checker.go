package infra

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// LancamentoExistenceChecker is the concrete implementation, in this same
// financeiro/infra package, of the read-only port 01-plano-de-contas defines
// for its renomear conta rule (FIN-D-008, PC-02 AC3): "01 never writes to
// 02, 02 never writes to 01" — it only ever reads lancamentos.
//
// Its interface is not declared here: 01's consumer (RenomearConta,
// 01-plano-de-contas/T4) will declare the narrow port it needs, and this
// type satisfies it structurally, the same AD-015 pattern used throughout
// this module. Wiring this into a financeiro composition root is deferred
// to whichever task first needs one (see financeiro/STATE.md) — no such
// root exists yet.
type LancamentoExistenceChecker struct{ db *gorm.DB }

func NewLancamentoExistenceChecker(db *gorm.DB) *LancamentoExistenceChecker {
	return &LancamentoExistenceChecker{db: db}
}

// TemLancamento reports whether contaID is referenced by at least one
// lançamento, regardless of status — including CANCELADA (PC-02 AC2).
func (c *LancamentoExistenceChecker) TemLancamento(ctx context.Context, contaID string) (bool, error) {
	var existe bool
	err := conn(ctx, c.db).Raw(`SELECT EXISTS (SELECT 1 FROM lancamentos WHERE conta_id = ?::uuid)`, contaID).
		Row().Scan(&existe)
	if err != nil {
		return false, fmt.Errorf("financeiro: verificar uso da conta: %w", err)
	}
	return existe, nil
}
