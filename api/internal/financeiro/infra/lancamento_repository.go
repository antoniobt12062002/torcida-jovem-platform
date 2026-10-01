package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
)

// LancamentoRepository persists and queries lancamentos.
type LancamentoRepository struct{ db *gorm.DB }

func NewLancamentoRepository(db *gorm.DB) *LancamentoRepository { return &LancamentoRepository{db: db} }

type lancamentoRow struct {
	ID                 string
	Tipo               string
	ContaID            string
	ValorBrutoCents    int64
	TaxaCents          int64
	ValorLiquidoCents  int64
	FormaPagamento     string
	Status             string
	DevolucaoDeID      *string
	MotivoCancelamento *string
	CanceladoPor       *string
	CanceladoEm        *time.Time
	CriadoPor          string
	CriadoEm           time.Time
	AtualizadoEm       time.Time
}

func (r lancamentoRow) lancamento() domain.Lancamento {
	return domain.Lancamento{
		ID: r.ID, Tipo: domain.TipoLancamento(r.Tipo), ContaID: r.ContaID,
		ValorBrutoCents: r.ValorBrutoCents, TaxaCents: r.TaxaCents, ValorLiquidoCents: r.ValorLiquidoCents,
		FormaPagamento: domain.FormaPagamento(r.FormaPagamento), Status: domain.StatusLancamento(r.Status),
		DevolucaoDeID: r.DevolucaoDeID, MotivoCancelamento: r.MotivoCancelamento,
		CanceladoPor: r.CanceladoPor, CanceladoEm: r.CanceladoEm,
		CriadoPor: r.CriadoPor, CriadoEm: r.CriadoEm, AtualizadoEm: r.AtualizadoEm,
	}
}

const lancamentoColumns = `id::text AS id, tipo, conta_id::text AS conta_id, valor_bruto_cents, taxa_cents,
	valor_liquido_cents, forma_pagamento, status, devolucao_de_id::text AS devolucao_de_id,
	motivo_cancelamento, cancelado_por::text AS cancelado_por, cancelado_em,
	criado_por::text AS criado_por, criado_em, atualizado_em`

func scanLancamento(row *sql.Row) (domain.Lancamento, error) {
	var lr lancamentoRow
	err := row.Scan(&lr.ID, &lr.Tipo, &lr.ContaID, &lr.ValorBrutoCents, &lr.TaxaCents,
		&lr.ValorLiquidoCents, &lr.FormaPagamento, &lr.Status, &lr.DevolucaoDeID,
		&lr.MotivoCancelamento, &lr.CanceladoPor, &lr.CanceladoEm,
		&lr.CriadoPor, &lr.CriadoEm, &lr.AtualizadoEm)
	if err != nil {
		return domain.Lancamento{}, err
	}
	return lr.lancamento(), nil
}

// Criar persists a new lançamento, born CRIADA.
func (r *LancamentoRepository) Criar(ctx context.Context, l domain.Lancamento) (domain.Lancamento, error) {
	row := conn(ctx, r.db).Raw(`INSERT INTO lancamentos
		(tipo, conta_id, valor_bruto_cents, taxa_cents, valor_liquido_cents, forma_pagamento, devolucao_de_id, criado_por)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING `+lancamentoColumns,
		string(l.Tipo), l.ContaID, l.ValorBrutoCents, l.TaxaCents, l.ValorLiquidoCents,
		string(l.FormaPagamento), nullableUUID(l.DevolucaoDeID), l.CriadoPor,
	).Row()
	lancamento, err := scanLancamento(row)
	if err != nil {
		return domain.Lancamento{}, fmt.Errorf("financeiro: criar lançamento: %w", err)
	}
	return lancamento, nil
}

// Buscar returns domain.ErrLancamentoNaoEncontrado when id does not exist.
func (r *LancamentoRepository) Buscar(ctx context.Context, id string) (domain.Lancamento, error) {
	row := conn(ctx, r.db).Raw(`SELECT `+lancamentoColumns+` FROM lancamentos WHERE id = ?::uuid`, id).Row()
	lancamento, err := scanLancamento(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Lancamento{}, domain.ErrLancamentoNaoEncontrado
		}
		return domain.Lancamento{}, fmt.Errorf("financeiro: buscar lançamento: %w", err)
	}
	return lancamento, nil
}

// Atualizar persists conta_id, valor_bruto_cents, taxa_cents,
// valor_liquido_cents and forma_pagamento (LAN-02 AC1). Whether the
// lançamento is still CRIADA is checked by the application (FIN-D-007), not
// here — same convention as the other cross-row business rules in this
// module. Its caller (EditarLancamento) always calls Buscar first in the
// same use case, and lancamentos are never deleted (no DELETE grant), so id
// not matching any row is not a reachable outcome here — unlike Buscar,
// there is no ErrLancamentoNaoEncontrado translation to test.
func (r *LancamentoRepository) Atualizar(ctx context.Context, l domain.Lancamento) (domain.Lancamento, error) {
	row := conn(ctx, r.db).Raw(`UPDATE lancamentos SET
		conta_id = ?, valor_bruto_cents = ?, taxa_cents = ?, valor_liquido_cents = ?, forma_pagamento = ?, atualizado_em = now()
		WHERE id = ?::uuid
		RETURNING `+lancamentoColumns,
		l.ContaID, l.ValorBrutoCents, l.TaxaCents, l.ValorLiquidoCents, string(l.FormaPagamento), l.ID,
	).Row()
	lancamento, err := scanLancamento(row)
	if err != nil {
		return domain.Lancamento{}, fmt.Errorf("financeiro: atualizar lançamento: %w", err)
	}
	return lancamento, nil
}

// Listar returns every lançamento, oldest first, with every field including
// status and devolucao_de_id (LAN-04 AC1) — nothing is filtered by default.
func (r *LancamentoRepository) Listar(ctx context.Context) ([]domain.Lancamento, error) {
	var rows []lancamentoRow
	err := conn(ctx, r.db).Raw(`SELECT ` + lancamentoColumns + ` FROM lancamentos ORDER BY criado_em`).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("financeiro: listar lançamentos: %w", err)
	}
	lancamentos := make([]domain.Lancamento, 0, len(rows))
	for _, row := range rows {
		lancamentos = append(lancamentos, row.lancamento())
	}
	return lancamentos, nil
}
