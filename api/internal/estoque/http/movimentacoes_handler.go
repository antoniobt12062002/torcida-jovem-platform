package estoquehttp

import (
	"context"

	estoqueapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/estoque/domain"
)

func movimentacaoOut(m domain.Movimentacao) Movimentacao {
	out := Movimentacao{
		Id: parseUUID(m.ID), ProdutoId: parseUUID(m.ProdutoID), Tipo: TipoMovimentacao(m.Tipo),
		Quantidade: m.Quantidade, Origem: OrigemMovimentacao(m.Origem), ResponsavelId: parseUUID(m.ResponsavelID),
		CriadoEm: m.CriadoEm.UTC(), Motivo: m.Motivo,
	}
	if m.MovimentacaoDeID != nil {
		id := parseUUID(*m.MovimentacaoDeID)
		out.MovimentacaoDeId = &id
	}
	return out
}

func (h *Handler) CreateMovimentacao(ctx context.Context, req CreateMovimentacaoRequestObject) (CreateMovimentacaoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	criada, err := h.M.RegistrarMovimentacao.Execute(requestContext(ctx), estoqueapp.RegistrarMovimentacaoInput{
		Actor: info.Principal, Tipo: domain.TipoMovimentacao(req.Body.Tipo), ProdutoID: req.Body.ProdutoId.String(),
		Quantidade: req.Body.Quantidade, Origem: domain.OrigemMovimentacao(req.Body.Origem),
		MovimentacaoDeID: uuidPtrToString(req.Body.MovimentacaoDeId),
	})
	if err != nil {
		return nil, err
	}
	return CreateMovimentacao201JSONResponse(movimentacaoOut(criada)), nil
}

func (h *Handler) ListMovimentacoes(ctx context.Context, req ListMovimentacoesRequestObject) (ListMovimentacoesResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	movimentacoes, err := h.M.ListarMovimentacoes.Execute(requestContext(ctx), estoqueapp.ListarMovimentacoesInput{
		Actor: info.Principal, ProdutoID: req.Id.String(),
	})
	if err != nil {
		return nil, err
	}
	out := MovimentacaoList{Items: make([]Movimentacao, len(movimentacoes))}
	for i, m := range movimentacoes {
		out.Items[i] = movimentacaoOut(m)
	}
	return ListMovimentacoes200JSONResponse(out), nil
}
