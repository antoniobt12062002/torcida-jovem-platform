// Package infra holds the financeiro repositories, implementing the ports
// each app use case defines.
package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// conn returns the transaction carried by ctx (database.WithTx), or the pool
// when there is none.
func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := database.TxFrom(ctx); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

// ContaRepository persists and queries contas_contabeis.
type ContaRepository struct{ db *gorm.DB }

func NewContaRepository(db *gorm.DB) *ContaRepository { return &ContaRepository{db: db} }

type contaRow struct {
	ID        string
	Tipo      string
	Nome      string
	ParentID  *string
	Ativo     bool
	CreatedAt time.Time
}

func (r contaRow) conta() domain.Conta {
	return domain.Conta{
		ID: r.ID, Tipo: domain.TipoConta(r.Tipo), Nome: r.Nome,
		ParentID: r.ParentID, Ativo: r.Ativo, CreatedAt: r.CreatedAt,
	}
}

const contaColumns = `id::text AS id, tipo, nome, parent_id::text AS parent_id, ativo, created_at`

// Buscar returns domain.ErrContaNaoEncontrada when id does not exist.
func (r *ContaRepository) Buscar(ctx context.Context, id string) (domain.Conta, error) {
	var row contaRow
	err := conn(ctx, r.db).Raw(`SELECT `+contaColumns+` FROM contas_contabeis WHERE id = ?::uuid`, id).
		Row().Scan(&row.ID, &row.Tipo, &row.Nome, &row.ParentID, &row.Ativo, &row.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Conta{}, domain.ErrContaNaoEncontrada
		}
		return domain.Conta{}, fmt.Errorf("financeiro: buscar conta: %w", err)
	}
	return row.conta(), nil
}

func (r *ContaRepository) Criar(ctx context.Context, c domain.Conta) (domain.Conta, error) {
	var row contaRow
	err := conn(ctx, r.db).Raw(`INSERT INTO contas_contabeis (tipo, nome, parent_id)
		VALUES (?, ?, ?) RETURNING `+contaColumns,
		string(c.Tipo), c.Nome, nullableUUID(c.ParentID),
	).Row().Scan(&row.ID, &row.Tipo, &row.Nome, &row.ParentID, &row.Ativo, &row.CreatedAt)
	if err != nil {
		return domain.Conta{}, fmt.Errorf("financeiro: criar conta: %w", err)
	}
	return row.conta(), nil
}

func (r *ContaRepository) Listar(ctx context.Context) ([]domain.Conta, error) {
	var rows []contaRow
	err := conn(ctx, r.db).Raw(`SELECT ` + contaColumns + ` FROM contas_contabeis ORDER BY nome`).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("financeiro: listar contas: %w", err)
	}
	contas := make([]domain.Conta, 0, len(rows))
	for _, row := range rows {
		contas = append(contas, row.conta())
	}
	return contas, nil
}

// Desativar sets ativo = false and returns domain.ErrContaNaoEncontrada when
// id does not exist. Deactivating an already-inactive conta matches the same
// WHERE id clause and succeeds again (idempotent, PC-03 Edge Cases).
func (r *ContaRepository) Desativar(ctx context.Context, id string) error {
	var got string
	err := conn(ctx, r.db).Raw(`UPDATE contas_contabeis SET ativo = false WHERE id = ?::uuid RETURNING id::text`, id).
		Row().Scan(&got)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrContaNaoEncontrada
		}
		return fmt.Errorf("financeiro: desativar conta: %w", err)
	}
	return nil
}

func nullableUUID(id *string) any {
	if id == nil {
		return nil
	}
	return *id
}
