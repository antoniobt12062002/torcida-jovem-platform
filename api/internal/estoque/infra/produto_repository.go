// Package infra holds the estoque repositories, implementing the ports each
// app use case defines.
package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
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

// ProdutoRepository persists and queries produtos_estoque.
type ProdutoRepository struct{ db *gorm.DB }

func NewProdutoRepository(db *gorm.DB) *ProdutoRepository { return &ProdutoRepository{db: db} }

type produtoRow struct {
	ID            string
	Codigo        string
	Nome          string
	UnidadeMedida string
	CriadoEm      time.Time
}

func (r produtoRow) produto() domain.Produto {
	return domain.Produto{ID: r.ID, Codigo: r.Codigo, Nome: r.Nome, UnidadeMedida: r.UnidadeMedida, CriadoEm: r.CriadoEm}
}

const produtoColumns = `id::text AS id, codigo, nome, unidade_medida, criado_em`

// Buscar returns domain.ErrProdutoNaoEncontrado when id does not exist.
func (r *ProdutoRepository) Buscar(ctx context.Context, id string) (domain.Produto, error) {
	var row produtoRow
	err := conn(ctx, r.db).Raw(`SELECT `+produtoColumns+` FROM produtos_estoque WHERE id = ?::uuid`, id).
		Row().Scan(&row.ID, &row.Codigo, &row.Nome, &row.UnidadeMedida, &row.CriadoEm)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Produto{}, domain.ErrProdutoNaoEncontrado
		}
		return domain.Produto{}, fmt.Errorf("estoque: buscar produto: %w", err)
	}
	return row.produto(), nil
}

// Criar inserts a produto, translating a duplicate codigo into
// domain.ErrCodigoDuplicado — never the raw driver error.
func (r *ProdutoRepository) Criar(ctx context.Context, p domain.Produto) (domain.Produto, error) {
	var row produtoRow
	err := conn(ctx, r.db).Raw(`INSERT INTO produtos_estoque (codigo, nome, unidade_medida)
		VALUES (?, ?, ?) RETURNING `+produtoColumns,
		p.Codigo, p.Nome, p.UnidadeMedida,
	).Row().Scan(&row.ID, &row.Codigo, &row.Nome, &row.UnidadeMedida, &row.CriadoEm)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" && pgErr.ConstraintName == "produtos_estoque_codigo_key" {
			return domain.Produto{}, domain.ErrCodigoDuplicado
		}
		return domain.Produto{}, fmt.Errorf("estoque: criar produto: %w", err)
	}
	return row.produto(), nil
}

func (r *ProdutoRepository) Listar(ctx context.Context) ([]domain.Produto, error) {
	var rows []produtoRow
	err := conn(ctx, r.db).Raw(`SELECT ` + produtoColumns + ` FROM produtos_estoque ORDER BY codigo`).Scan(&rows).Error
	if err != nil {
		return nil, fmt.Errorf("estoque: listar produtos: %w", err)
	}
	produtos := make([]domain.Produto, 0, len(rows))
	for _, row := range rows {
		produtos = append(produtos, row.produto())
	}
	return produtos, nil
}
