package domain

import (
	"errors"
	"time"
)

// TipoLancamento is the same RECEITA/DESPESA concept as TipoConta (FIN-D-015
// requires a lançamento's tipo to match its conta's tipo, so both use the
// same underlying type — no conversion needed to compare them).
type TipoLancamento = TipoConta

// StatusLancamento is the lifecycle state of a Lancamento. Only CRIADA is
// reachable from this sub-spec (01-plano-de-contas, 02-lancamentos); the
// CRIADA→RECEBIDA/PAGA/CANCELADA transitions belong to 03-workflow-e-saldo.
type StatusLancamento string

const (
	StatusCriada    StatusLancamento = "CRIADA"
	StatusRecebida  StatusLancamento = "RECEBIDA"
	StatusPaga      StatusLancamento = "PAGA"
	StatusCancelada StatusLancamento = "CANCELADA"
)

// FormaPagamento is how a lançamento's money moved.
type FormaPagamento string

const (
	FormaPix           FormaPagamento = "PIX"
	FormaCartao        FormaPagamento = "CARTAO"
	FormaDinheiro      FormaPagamento = "DINHEIRO"
	FormaTransferencia FormaPagamento = "TRANSFERENCIA"
	FormaOutros        FormaPagamento = "OUTROS"
)

// Lancamento is a receita or despesa (LAN-01). ValorLiquidoCents is computed
// once at creation (ValorBrutoCents - TaxaCents) and never recomputed.
// DevolucaoDeID, when set, points to the RECEITA this despesa returns
// (FIN-D-003) — unidirectional: the referenced receita carries no field back.
type Lancamento struct {
	ID                 string
	Tipo               TipoLancamento
	ContaID            string
	ValorBrutoCents    int64
	TaxaCents          int64
	ValorLiquidoCents  int64
	FormaPagamento     FormaPagamento
	Status             StatusLancamento
	DevolucaoDeID      *string
	MotivoCancelamento *string
	CanceladoPor       *string
	CanceladoEm        *time.Time
	CriadoPor          string
	CriadoEm           time.Time
	AtualizadoEm       time.Time
}

var (
	// ErrContaInvalida is returned when a lançamento's conta_id does not
	// exist or refers to an inactive conta (LAN-01 AC2).
	ErrContaInvalida = errors.New("conta inválida")
	// ErrLancamentoTipoIncompativel is returned when a lançamento's tipo does
	// not match its conta's tipo (LAN-01 AC3, FIN-D-015; error code
	// conta_tipo_incompativel per the spec — a distinct Go sentinel from
	// conta.go's ErrContaTipoIncompativel, which guards PC-01's parent/child
	// consistency, a different rule).
	ErrLancamentoTipoIncompativel = errors.New("tipo do lançamento incompatível com o tipo da conta")
	// ErrLancamentoNaoEncontrado is returned when a lançamento id does not
	// exist.
	ErrLancamentoNaoEncontrado = errors.New("lançamento não encontrado")
	// ErrLancamentoImutavel is returned when editing a lançamento outside
	// status CRIADA (LAN-02 AC2).
	ErrLancamentoImutavel = errors.New("lançamento não pode mais ser editado")
	// ErrDevolucaoInvalida is returned when a devolução's referenced
	// lançamento is not a RECEITA RECEBIDA (LAN-03 AC2).
	ErrDevolucaoInvalida = errors.New("devolução inválida")
	// ErrLancamentoNaoPodeSerRecebido is returned when receiving a
	// lançamento that is not a CRIADA RECEITA (WKF-01 AC2, Edge Cases).
	ErrLancamentoNaoPodeSerRecebido = errors.New("lançamento não pode ser recebido")
	// ErrLancamentoNaoPodeSerPago is returned when paying a lançamento that
	// is not a CRIADA DESPESA (WKF-01 AC4, Edge Cases).
	ErrLancamentoNaoPodeSerPago = errors.New("lançamento não pode ser pago")
)
