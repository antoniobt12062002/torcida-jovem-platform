package infra

import (
	"context"
	"database/sql"
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
