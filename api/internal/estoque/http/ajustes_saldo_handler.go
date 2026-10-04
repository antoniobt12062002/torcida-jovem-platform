package estoquehttp

import (
	"context"

	estoqueapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
)

func (h *Handler) CreateAjuste(ctx context.Context, req CreateAjusteRequestObject) (CreateAjusteResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	criada, err := h.M.AjustarEstoque.Execute(requestContext(ctx), estoqueapp.AjustarEstoqueInput{
		Actor: info.Principal, ProdutoID: req.Body.ProdutoId.String(), Quantidade: req.Body.Quantidade, Motivo: req.Body.Motivo,
	})
	if err != nil {
		return nil, err
	}
	return CreateAjuste201JSONResponse(movimentacaoOut(criada)), nil
}

func (h *Handler) GetSaldo(ctx context.Context, req GetSaldoRequestObject) (GetSaldoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	saldo, err := h.M.ConsultarSaldo.Execute(requestContext(ctx), estoqueapp.ConsultarSaldoInput{
		Actor: info.Principal, ProdutoID: req.Id.String(),
	})
	if err != nil {
		return nil, err
	}
	return GetSaldo200JSONResponse(Saldo{Saldo: saldo}), nil
}
