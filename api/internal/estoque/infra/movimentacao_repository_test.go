//go:build integration

package infra_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/database"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

func newTestProduto(t *testing.T, repo *infra.ProdutoRepository) string {
	t.Helper()
	p, err := repo.Criar(context.Background(), domain.Produto{Codigo: uuid.NewString(), Nome: "x", UnidadeMedida: "UN"})
	if err != nil {
		t.Fatalf("criar produto de teste: %v", err)
	}
	return p.ID
}

func newTestUser(t *testing.T, db *gorm.DB) string {
	t.Helper()
	var id string
	err := db.Raw(`INSERT INTO users (email, name, password_hash) VALUES (?, 'Fulano', 'hash') RETURNING id::text`,
		uuid.NewString()+"@exemplo.com").Scan(&id).Error
	if err != nil || id == "" {
		t.Fatalf("criar usuário de teste: id=%q err=%v", id, err)
	}
	return id
}

func TestCriarPersistsAndBuscarReturnsIt(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)
	userID := newTestUser(t, db)
	ctx := context.Background()

	criada, err := movs.Criar(ctx, domain.Movimentacao{
		ProdutoID: produtoID, Tipo: domain.Entrada, Quantidade: 10, Origem: domain.OrigemCompra, ResponsavelID: userID,
	})
	if err != nil {
		t.Fatalf("criar: %v", err)
	}
	if criada.ID == "" || criada.Quantidade != 10 || criada.Tipo != domain.Entrada {
		t.Errorf("movimentação criada = %+v", criada)
	}

	buscada, err := movs.Buscar(ctx, criada.ID)
	if err != nil {
		t.Fatalf("buscar: %v", err)
	}
	if buscada.ID != criada.ID {
		t.Errorf("buscada = %+v", buscada)
	}
}

func TestBuscarReturnsErrMovimentacaoNaoEncontrada(t *testing.T) {
	db := testutil.NewTestDB(t)
	movs := infra.NewMovimentacaoRepository(db)

	_, err := movs.Buscar(context.Background(), uuid.NewString())

	if !errors.Is(err, domain.ErrMovimentacaoNaoEncontrada) {
		t.Errorf("err = %v, esperado ErrMovimentacaoNaoEncontrada", err)
	}
}

func TestListarPorProdutoReturnsInCreationOrder(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)
	userID := newTestUser(t, db)
	ctx := context.Background()

	primeira, err := movs.Criar(ctx, domain.Movimentacao{ProdutoID: produtoID, Tipo: domain.Entrada, Quantidade: 10, Origem: domain.OrigemCompra, ResponsavelID: userID})
	if err != nil {
		t.Fatalf("criar primeira: %v", err)
	}
	segunda, err := movs.Criar(ctx, domain.Movimentacao{ProdutoID: produtoID, Tipo: domain.Saida, Quantidade: -3, Origem: domain.OrigemVenda, ResponsavelID: userID})
	if err != nil {
		t.Fatalf("criar segunda: %v", err)
	}

	lista, err := movs.ListarPorProduto(ctx, produtoID)
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(lista) != 2 || lista[0].ID != primeira.ID || lista[1].ID != segunda.ID {
		t.Errorf("lista = %+v", lista)
	}
}

func TestSaldoComLockReflectsTheSumOfMovimentacoes(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)
	userID := newTestUser(t, db)
	ctx := context.Background()

	if _, err := movs.Criar(ctx, domain.Movimentacao{ProdutoID: produtoID, Tipo: domain.Entrada, Quantidade: 10, Origem: domain.OrigemCompra, ResponsavelID: userID}); err != nil {
		t.Fatalf("criar entrada: %v", err)
	}
	if _, err := movs.Criar(ctx, domain.Movimentacao{ProdutoID: produtoID, Tipo: domain.Saida, Quantidade: -4, Origem: domain.OrigemVenda, ResponsavelID: userID}); err != nil {
		t.Fatalf("criar saída: %v", err)
	}

	err := database.WithTx(ctx, db, func(ctx context.Context) error {
		saldo, err := movs.SaldoComLock(ctx, produtoID)
		if err != nil {
			return err
		}
		if saldo != 6 {
			t.Errorf("saldo = %d, esperado 6", saldo)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
}

func TestSaldoComLockReturnsZeroForAProdutoWithNoMovimentacoes(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)

	var saldo int64
	err := database.WithTx(context.Background(), db, func(ctx context.Context) error {
		var err error
		saldo, err = movs.SaldoComLock(ctx, produtoID)
		return err
	})
	if err != nil {
		t.Fatalf("tx: %v", err)
	}
	if saldo != 0 {
		t.Errorf("saldo = %d, esperado 0", saldo)
	}
}

// Teste obrigatório de concorrência: saldo = 1, duas goroutines tentam
// saída de 1 cada. Exatamente uma sucede, a outra recebe saldo_insuficiente
// (verificado pelo próprio chamador da checagem — aqui simulado dentro da
// mesma transação de cada goroutine, exatamente como RegistrarMovimentacao
// fará em 02-movimentacoes/T4). Saldo final nunca fica negativo.
func TestConcurrentSaidasDisputingTheLastUnitNeverBothSucceed(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)
	userID := newTestUser(t, db)
	ctx := context.Background()

	if _, err := movs.Criar(ctx, domain.Movimentacao{ProdutoID: produtoID, Tipo: domain.Entrada, Quantidade: 1, Origem: domain.OrigemCompra, ResponsavelID: userID}); err != nil {
		t.Fatalf("criar entrada inicial: %v", err)
	}

	trySaida := func() error {
		return database.WithTx(context.Background(), db, func(ctx context.Context) error {
			saldo, err := movs.SaldoComLock(ctx, produtoID)
			if err != nil {
				return err
			}
			if saldo-1 < 0 {
				return domain.ErrSaldoInsuficiente
			}
			_, err = movs.Criar(ctx, domain.Movimentacao{ProdutoID: produtoID, Tipo: domain.Saida, Quantidade: -1, Origem: domain.OrigemVenda, ResponsavelID: userID})
			return err
		})
	}

	var wg sync.WaitGroup
	results := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			results[i] = trySaida()
		}(i)
	}
	wg.Wait()

	successes, insufficient := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrSaldoInsuficiente):
			insufficient++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if successes != 1 || insufficient != 1 {
		t.Errorf("sucessos=%d insuficientes=%d, esperado 1 e 1", successes, insufficient)
	}

	var saldoFinal int64
	err := database.WithTx(ctx, db, func(ctx context.Context) error {
		var err error
		saldoFinal, err = movs.SaldoComLock(ctx, produtoID)
		return err
	})
	if err != nil {
		t.Fatalf("tx final: %v", err)
	}
	if saldoFinal != 0 {
		t.Errorf("saldo final = %d, esperado 0 (nunca negativo)", saldoFinal)
	}
}

// Prova determinística do mecanismo (independente de sorte de agendamento
// do SO, que o teste de outcome acima sozinho não garante contra um
// mutante que remova o lock mas mantenha o SELECT rápido o suficiente para
// nunca colidir de verdade): uma segunda chamada a SaldoComLock do mesmo
// produto_id, numa transação separada, fica bloqueada enquanto a primeira
// mantém sua transação aberta, e só desbloqueia depois que a primeira
// comita (libera o pg_advisory_xact_lock).
func TestSaldoComLockSerializesConcurrentCallersOfTheSameProduto(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoID := newTestProduto(t, produtos)

	aHasLock := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = database.WithTx(context.Background(), db, func(ctx context.Context) error {
			if _, err := movs.SaldoComLock(ctx, produtoID); err != nil {
				return err
			}
			close(aHasLock)
			<-release
			return nil
		})
	}()

	<-aHasLock
	bDone := make(chan struct{})
	start := time.Now()
	go func() {
		_ = database.WithTx(context.Background(), db, func(ctx context.Context) error {
			_, err := movs.SaldoComLock(ctx, produtoID)
			return err
		})
		close(bDone)
	}()

	select {
	case <-bDone:
		t.Fatal("B deveria estar bloqueada esperando o lock de A, mas retornou antes de A liberar")
	case <-time.After(150 * time.Millisecond):
	}

	close(release)
	select {
	case <-bDone:
	case <-time.After(2 * time.Second):
		t.Fatal("B deveria desbloquear depois que A comitou (liberou o advisory lock)")
	}
	if elapsed := time.Since(start); elapsed < 150*time.Millisecond {
		t.Errorf("B desbloqueou cedo demais (%v): o lock não está serializando as chamadas", elapsed)
	}
}

// O lock é escopado por produto_id (hashtext(produto_id)), nunca global:
// duas chamadas para SKUs diferentes nunca se bloqueiam uma à outra.
func TestSaldoComLockDoesNotSerializeCallersOfDifferentProdutos(t *testing.T) {
	db := testutil.NewTestDB(t)
	produtos := infra.NewProdutoRepository(db)
	movs := infra.NewMovimentacaoRepository(db)
	produtoA := newTestProduto(t, produtos)
	produtoB := newTestProduto(t, produtos)

	aHasLock := make(chan struct{})
	release := make(chan struct{})
	go func() {
		_ = database.WithTx(context.Background(), db, func(ctx context.Context) error {
			if _, err := movs.SaldoComLock(ctx, produtoA); err != nil {
				return err
			}
			close(aHasLock)
			<-release
			return nil
		})
	}()

	<-aHasLock
	bDone := make(chan struct{})
	go func() {
		_ = database.WithTx(context.Background(), db, func(ctx context.Context) error {
			_, err := movs.SaldoComLock(ctx, produtoB)
			return err
		})
		close(bDone)
	}()

	select {
	case <-bDone:
	case <-time.After(1 * time.Second):
		t.Fatal("B (produto diferente) não deveria esperar pelo lock de A")
	}
	close(release)
}
