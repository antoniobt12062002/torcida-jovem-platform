//go:build integration

package app_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
)

// txRecordingStore is the real repository, recording which transaction each
// call ran in — so a test can prove the lock and the insert share one.
type txRecordingStore struct {
	*infra.MovimentacaoRepository
	mu         sync.Mutex
	lockTx     *gorm.DB
	lockInTx   bool
	lockCalled bool
	criarTx    *gorm.DB
}

func (s *txRecordingStore) SaldoComLock(ctx context.Context, produtoID string) (int64, error) {
	s.mu.Lock()
	s.lockTx, s.lockInTx = database.TxFrom(ctx)
	s.lockCalled = true
	s.mu.Unlock()
	return s.MovimentacaoRepository.SaldoComLock(ctx, produtoID)
}

func (s *txRecordingStore) Criar(ctx context.Context, m domain.Movimentacao) (domain.Movimentacao, error) {
	s.mu.Lock()
	s.criarTx, _ = database.TxFrom(ctx)
	s.mu.Unlock()
	return s.MovimentacaoRepository.Criar(ctx, m)
}

// MOV-02 AC5 (AD-009): a checagem de saldo de uma saída roda com o lock
// dentro da MESMA transação do insert — senão o pg_advisory_xact_lock é
// liberado no fim do próprio statement e a corrida pela última unidade
// reabre. Prova determinística, independente de agendamento de goroutines.
func TestRegistrarMovimentacaoSaidaLocksTheBalanceInsideTheSameTransactionAsTheInsert(t *testing.T) {
	e := newEnv(t)
	produtoID := e.criarProdutoDeTeste(t, "LOCK-TX")
	actorRM := e.userActor(t, app.PermMovimentacaoCreate)
	if _, err := e.registrarMovimentacao().Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 1, Origem: domain.OrigemCompra,
	}); err != nil {
		t.Fatalf("entrada: %v", err)
	}
	store := &txRecordingStore{MovimentacaoRepository: e.movimentacoes}
	uc := &app.RegistrarMovimentacao{Authz: simpleAuthz{}, Produtos: e.produtos, Movimentacoes: store, Audit: e.rec, Tx: e.tx}

	if _, err := uc.Execute(context.Background(), app.RegistrarMovimentacaoInput{
		Actor: actorRM, Tipo: domain.Saida, ProdutoID: produtoID, Quantidade: 1, Origem: domain.OrigemVenda,
	}); err != nil {
		t.Fatalf("saída: %v", err)
	}

	if !store.lockCalled {
		t.Fatal("uma saída precisa consultar o saldo com lock")
	}
	if !store.lockInTx {
		t.Fatal("SaldoComLock rodou fora de uma transação: o advisory lock seria liberado antes do insert")
	}
	if store.lockTx != store.criarTx {
		t.Error("SaldoComLock e Criar rodaram em transações diferentes: a checagem não é atômica com o insert")
	}
}

// MOV-02 AC5, o cenário literal exigido pelo mantenedor, agora pelo caso de
// uso real (não pelo repositório): saldo=1, duas saídas de 1 concorrentes —
// exatamente uma sucede, a outra recebe saldo_insuficiente, saldo final 0.
// Repetido em várias rodadas para aumentar a chance de colisão real.
func TestConcurrentSaidasThroughTheUseCaseNeverBothTakeTheLastUnit(t *testing.T) {
	e := newEnv(t)
	actorRM := e.userActor(t, app.PermMovimentacaoCreate, app.PermSaldoRead)
	uc := e.registrarMovimentacao()

	for round := range 10 {
		produtoID := e.criarProdutoDeTeste(t, "CONC-UC-"+string(rune('A'+round)))
		if _, err := uc.Execute(context.Background(), app.RegistrarMovimentacaoInput{
			Actor: actorRM, Tipo: domain.Entrada, ProdutoID: produtoID, Quantidade: 1, Origem: domain.OrigemCompra,
		}); err != nil {
			t.Fatalf("rodada %d, entrada: %v", round, err)
		}

		start := make(chan struct{})
		errs := make([]error, 2)
		var wg sync.WaitGroup
		for i := range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, errs[i] = uc.Execute(context.Background(), app.RegistrarMovimentacaoInput{
					Actor: actorRM, Tipo: domain.Saida, ProdutoID: produtoID, Quantidade: 1, Origem: domain.OrigemVenda,
				})
			}()
		}
		close(start)
		wg.Wait()

		var ok, insuficiente int
		for _, err := range errs {
			switch {
			case err == nil:
				ok++
			case errors.Is(err, domain.ErrSaldoInsuficiente):
				insuficiente++
			default:
				t.Fatalf("rodada %d: erro inesperado %v", round, err)
			}
		}
		if ok != 1 || insuficiente != 1 {
			t.Fatalf("rodada %d: %d sucessos e %d saldo_insuficiente, esperado exatamente 1 e 1", round, ok, insuficiente)
		}
		saldo, err := e.consultarSaldo().Execute(context.Background(), app.ConsultarSaldoInput{Actor: actorRM, ProdutoID: produtoID})
		if err != nil || saldo != 0 {
			t.Fatalf("rodada %d: saldo = %d, err = %v, esperado 0 (nunca negativo)", round, saldo, err)
		}
	}
}
