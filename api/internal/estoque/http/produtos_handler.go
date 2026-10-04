package estoquehttp

import (
	"context"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	estoqueapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
)

func parseUUID(id string) openapi_types.UUID {
	u, _ := uuid.Parse(id)
	return openapi_types.UUID(u)
}

func uuidPtrToString(u *openapi_types.UUID) *string {
	if u == nil {
		return nil
	}
	s := uuid.UUID(*u).String()
	return &s
}

func produtoOut(p domain.Produto) Produto {
	return Produto{
		Id: parseUUID(p.ID), Codigo: p.Codigo, Nome: p.Nome, UnidadeMedida: p.UnidadeMedida, CriadoEm: p.CriadoEm.UTC(),
	}
}

func (h *Handler) CreateProduto(ctx context.Context, req CreateProdutoRequestObject) (CreateProdutoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	criado, err := h.M.CriarProduto.Execute(requestContext(ctx), estoqueapp.CriarProdutoInput{
		Actor: info.Principal, Codigo: req.Body.Codigo, Nome: req.Body.Nome, UnidadeMedida: req.Body.UnidadeMedida,
	})
	if err != nil {
		return nil, err
	}
	return CreateProduto201JSONResponse(produtoOut(criado)), nil
}

func (h *Handler) ListProdutos(ctx context.Context, _ ListProdutosRequestObject) (ListProdutosResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	produtos, err := h.M.ListarProdutos.Execute(requestContext(ctx), estoqueapp.ListarProdutosInput{Actor: info.Principal})
	if err != nil {
		return nil, err
	}
	out := ProdutoList{Items: make([]Produto, len(produtos))}
	for i, p := range produtos {
		out.Items[i] = produtoOut(p)
	}
	return ListProdutos200JSONResponse(out), nil
}
