//go:build integration

package infra_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/google/uuid"

	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/infra"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/testutil"
)

func TestCriarPersistsAndListarReturnsIt(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := infra.NewProdutoRepository(db)
	ctx := context.Background()

	criado, err := repo.Criar(ctx, domain.Produto{Codigo: "X-1", Nome: "Produto X", UnidadeMedida: "UN"})
	if err != nil {
		t.Fatalf("criar: %v", err)
	}
	if criado.ID == "" || criado.Codigo != "X-1" || criado.Nome != "Produto X" || criado.UnidadeMedida != "UN" || criado.CriadoEm.IsZero() {
		t.Errorf("produto criado = %+v", criado)
	}

	lista, err := repo.Listar(ctx)
	if err != nil {
		t.Fatalf("listar: %v", err)
	}
	if len(lista) != 1 || lista[0].ID != criado.ID {
		t.Errorf("lista = %+v", lista)
	}
}

func TestBuscarReturnsErrProdutoNaoEncontrado(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := infra.NewProdutoRepository(db)

	_, err := repo.Buscar(context.Background(), uuid.NewString())

	if !errors.Is(err, domain.ErrProdutoNaoEncontrado) {
		t.Errorf("err = %v, esperado ErrProdutoNaoEncontrado", err)
	}
}

func TestCriarTranslatesDuplicateCodigoIntoErrCodigoDuplicado(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := infra.NewProdutoRepository(db)
	ctx := context.Background()
	if _, err := repo.Criar(ctx, domain.Produto{Codigo: "DUP", Nome: "a", UnidadeMedida: "UN"}); err != nil {
		t.Fatalf("criar primeiro: %v", err)
	}

	_, err := repo.Criar(ctx, domain.Produto{Codigo: "DUP", Nome: "b", UnidadeMedida: "UN"})

	if !errors.Is(err, domain.ErrCodigoDuplicado) {
		t.Errorf("err = %v, esperado ErrCodigoDuplicado", err)
	}
}

// Duas criações concorrentes do mesmo código: só uma conclui, a outra
// recebe ErrCodigoDuplicado — nunca as duas sucedem, nunca as duas falham
// com outro erro.
func TestConcurrentCreationsOfTheSameCodigoOnlyOneSucceeds(t *testing.T) {
	db := testutil.NewTestDB(t)
	repo := infra.NewProdutoRepository(db)

	var wg sync.WaitGroup
	results := make([]error, 2)
	for i := range 2 {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, results[i] = repo.Criar(context.Background(), domain.Produto{Codigo: "RACE", Nome: "x", UnidadeMedida: "UN"})
		}(i)
	}
	wg.Wait()

	successes, duplicates := 0, 0
	for _, err := range results {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrCodigoDuplicado):
			duplicates++
		default:
			t.Errorf("erro inesperado: %v", err)
		}
	}
	if successes != 1 || duplicates != 1 {
		t.Errorf("sucessos=%d duplicados=%d, esperado 1 e 1", successes, duplicates)
	}
}
