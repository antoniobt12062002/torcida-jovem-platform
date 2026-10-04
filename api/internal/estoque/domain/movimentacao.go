package domain

import (
	"errors"
	"time"
)

// TipoMovimentacao is one of the 4 stock movement types (AD-009/ADR-007).
type TipoMovimentacao string

const (
	Entrada   TipoMovimentacao = "ENTRADA"
	Saida     TipoMovimentacao = "SAIDA"
	Ajuste    TipoMovimentacao = "AJUSTE"
	Devolucao TipoMovimentacao = "DEVOLUCAO"
)

// OrigemMovimentacao is where a movement came from. VENDA, COMPRA and
// EVENTO are reserved values for future integrations with loja, financeiro
// and eventos — none of those modules call estoque yet (EST-D-002, "Fora do
// V1"); only INVENTARIO and AJUSTE_MANUAL have a real path in this V1.
type OrigemMovimentacao string

const (
	OrigemVenda        OrigemMovimentacao = "VENDA"
	OrigemCompra       OrigemMovimentacao = "COMPRA"
	OrigemEvento       OrigemMovimentacao = "EVENTO"
	OrigemInventario   OrigemMovimentacao = "INVENTARIO"
	OrigemAjusteManual OrigemMovimentacao = "AJUSTE_MANUAL"
)

// Movimentacao is a single, immutable stock ledger entry. Quantidade is
// already signed by the application (positive or negative, never zero) so
// the balance is a plain SUM(quantidade), never a CASE per tipo
// (design/02-movimentacoes.md).
type Movimentacao struct {
	ID               string
	ProdutoID        string
	Tipo             TipoMovimentacao
	Quantidade       int64
	Origem           OrigemMovimentacao
	Motivo           *string // only filled when Tipo == Ajuste
	MovimentacaoDeID *string // only filled when Tipo == Devolucao
	ResponsavelID    string
	CriadoEm         time.Time
}

var (
	// ErrMovimentacaoNaoEncontrada is returned when a movimentacao id does
	// not exist. Never surfaces to HTTP directly — RegistrarMovimentacao
	// translates it into ErrDevolucaoInvalida (same pattern financeiro's
	// CriarDevolucao already applies to ErrLancamentoNaoEncontrado).
	ErrMovimentacaoNaoEncontrada = errors.New("movimentação não encontrada")
	// ErrDevolucaoInvalida is returned when movimentacao_de_id does not
	// exist, references a different produto, or references an AJUSTE or
	// another DEVOLUCAO (EST-D-009).
	ErrDevolucaoInvalida = errors.New("devolução inválida")
	// ErrSaldoInsuficiente is returned when a SAIDA, or a DEVOLUCAO that
	// reverses an ENTRADA, would make the balance negative.
	ErrSaldoInsuficiente = errors.New("saldo insuficiente")
	// ErrQuantidadeInvalida is returned when quantidade is zero, or
	// negative where a positive value is expected (ENTRADA/SAIDA/DEVOLUCAO).
	ErrQuantidadeInvalida = errors.New("quantidade inválida")
	// ErrMotivoObrigatorio is returned when an AJUSTE's motivo is blank.
	ErrMotivoObrigatorio = errors.New("motivo é obrigatório")
)
