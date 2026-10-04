package domain

import (
	"errors"
	"time"
)

// Produto is the minimal estoque SKU: just enough to identify and move
// stock. It carries no commercial catalog data (price, description, image,
// category) — that belongs to a future loja module, which may reference
// this SKU by id (EST-D-001). It is stable after creation: there is no
// update, deactivate or delete case use (EST-D-007).
type Produto struct {
	ID            string
	Codigo        string
	Nome          string
	UnidadeMedida string
	CriadoEm      time.Time
}

var (
	// ErrProdutoNaoEncontrado is returned when a produto id does not exist.
	ErrProdutoNaoEncontrado = errors.New("produto não encontrado")
	// ErrCodigoDuplicado is returned when codigo is already used by another produto.
	ErrCodigoDuplicado = errors.New("código já utilizado")
	// ErrCampoObrigatorio is returned when codigo, nome or unidade_medida is blank (PRD-01 AC3).
	ErrCampoObrigatorio = errors.New("campo obrigatório")
)
