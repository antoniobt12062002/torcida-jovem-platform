package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
)

// MovimentacaoRepository persists and queries movimentacoes_estoque.
type MovimentacaoRepository struct{ db *gorm.DB }

func NewMovimentacaoRepository(db *gorm.DB) *MovimentacaoRepository {
	return &MovimentacaoRepository{db: db}
}

type movimentacaoRow struct {
	ID               string
	ProdutoID        string
	Tipo             string
	Quantidade       int64
	Origem           string
	Motivo           *string
	MovimentacaoDeID *string
	ResponsavelID    string
	CriadoEm         time.Time
}

func (r movimentacaoRow) movimentacao() domain.Movimentacao {
	return domain.Movimentacao{
		ID: r.ID, ProdutoID: r.ProdutoID, Tipo: domain.TipoMovimentacao(r.Tipo), Quantidade: r.Quantidade,
		Origem: domain.OrigemMovimentacao(r.Origem), Motivo: r.Motivo, MovimentacaoDeID: r.MovimentacaoDeID,
		ResponsavelID: r.ResponsavelID, CriadoEm: r.CriadoEm,
	}
}

const movimentacaoColumns = `id::text AS id, produto_id::text AS produto_id, tipo, quantidade, origem,
	motivo, movimentacao_de_id::text AS movimentacao_de_id, responsavel_id::text AS responsavel_id, criado_em`

// Buscar returns domain.ErrMovimentacaoNaoEncontrada when id does not exist.
func (r *MovimentacaoRepository) Buscar(ctx context.Context, id string) (domain.Movimentacao, error) {
	var row movimentacaoRow
	err := conn(ctx, r.db).Raw(`SELECT `+movimentacaoColumns+` FROM movimentacoes_estoque WHERE id = ?::uuid`, id).
		Row().Scan(&row.ID, &row.ProdutoID, &row.Tipo, &row.Quantidade, &row.Origem, &row.Motivo, &row.MovimentacaoDeID, &row.ResponsavelID, &row.CriadoEm)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Movimentacao{}, domain.ErrMovimentacaoNaoEncontrada
		}
		return domain.Movimentacao{}, fmt.Errorf("estoque: buscar movimentação: %w", err)
	}
	return row.movimentacao(), nil
}

func (r *MovimentacaoRepository) Criar(ctx context.Context, m domain.Movimentacao) (domain.Movimentacao, error) {
	var row movimentacaoRow
	err := conn(ctx, r.db).Raw(`INSERT INTO movimentacoes_estoque
		(produto_id, tipo, quantidade, origem, motivo, movimentacao_de_id, responsavel_id)
		VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING `+movimentacaoColumns,
		m.ProdutoID, string(m.Tipo), m.Quantidade, string(m.Origem), m.Motivo, nullableUUID(m.MovimentacaoDeID), m.ResponsavelID,
	).Row().Scan(&row.ID, &row.ProdutoID, &row.Tipo, &row.Quantidade, &row.Origem, &row.Motivo, &row.MovimentacaoDeID, &row.ResponsavelID, &row.CriadoEm)
	if err != nil {
		return domain.Movimentacao{}, fmt.Errorf("estoque: criar movimentação: %w", err)
	}
	return row.movimentacao(), nil
}

func (r *MovimentacaoRepository) ListarPorProduto(ctx context.Context, produtoID string) ([]domain.Movimentacao, error) {
	var rows []movimentacaoRow
	err := conn(ctx, r.db).Raw(`SELECT `+movimentacaoColumns+` FROM movimentacoes_estoque WHERE produto_id = ?::uuid ORDER BY criado_em`, produtoID).
		Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("estoque: listar movimentações: %w", err)
	}
	out := make([]domain.Movimentacao, 0, len(rows))
	for _, row := range rows {
		out = append(out, row.movimentacao())
	}
	return out, nil
}

// SaldoComLock takes a Postgres advisory lock scoped to produtoID
// (pg_advisory_xact_lock, released automatically at the end of the caller's
// transaction) before computing SUM(quantidade). It must always run inside
// the same transaction as the Criar call that follows it — this serializes
// every negative-effect write for the same produto, so two concurrent
// saídas disputing the last unit never both succeed (design/
// 02-movimentacoes.md's Tech Decisions: FOR UPDATE does not work with
// SUM(), so a logical-key advisory lock is used instead, same pattern
// already in identity/infra).
func (r *MovimentacaoRepository) SaldoComLock(ctx context.Context, produtoID string) (int64, error) {
	tx := conn(ctx, r.db)
	if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", produtoID).Error; err != nil {
		return 0, fmt.Errorf("estoque: lock de saldo: %w", err)
	}
	var saldo int64
	err := tx.Raw(`SELECT COALESCE(SUM(quantidade), 0) FROM movimentacoes_estoque WHERE produto_id = ?::uuid`, produtoID).
		Scan(&saldo).Error
	if err != nil {
		return 0, fmt.Errorf("estoque: calcular saldo: %w", err)
	}
	return saldo, nil
}

// Saldo (sem lock, para ConsultarSaldo) é adicionado em 03-ajustes-e-saldo/T2
// — não antes, para manter os limites de tarefa aprovados no DAG.

func nullableUUID(id *string) any {
	if id == nil {
		return nil
	}
	return *id
}
