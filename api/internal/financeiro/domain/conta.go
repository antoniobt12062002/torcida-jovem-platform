// Package domain holds the financeiro entities and rules, free of
// infrastructure.
package domain

import (
	"errors"
	"time"
)

// TipoConta is RECEITA or DESPESA.
type TipoConta string

const (
	TipoReceita TipoConta = "RECEITA"
	TipoDespesa TipoConta = "DESPESA"
)

// Conta is one node of the plano de contas (PC-01). Its Nome is mutable only
// until the first lançamento references it (FIN-D-008, PC-02); Ativo, never
// deletion, marks it retired (PC-03).
type Conta struct {
	ID        string
	Tipo      TipoConta
	Nome      string
	ParentID  *string
	Ativo     bool
	CreatedAt time.Time
}

var (
	// ErrContaNaoEncontrada is returned when a conta id (including a
	// parent_id given at creation) does not exist.
	ErrContaNaoEncontrada = errors.New("conta não encontrada")
	// ErrContaTipoIncompativel: a subconta must have the same Tipo as its
	// parent (PC-01 AC2).
	ErrContaTipoIncompativel = errors.New("conta filha deve ter o mesmo tipo do pai")
)
