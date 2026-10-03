package financeirohttp

import (
	"context"

	finapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/platform/money"
)

func lancamentoOut(l domain.Lancamento) Lancamento {
	out := Lancamento{
		Id: parseUUID(l.ID), Tipo: TipoConta(l.Tipo), ContaId: parseUUID(l.ContaID),
		ValorBrutoCents: money.Cents(l.ValorBrutoCents), TaxaCents: money.Cents(l.TaxaCents), ValorLiquidoCents: money.Cents(l.ValorLiquidoCents),
		FormaPagamento: FormaPagamento(l.FormaPagamento), Status: StatusLancamento(l.Status),
		CriadoPor: parseUUID(l.CriadoPor), CriadoEm: l.CriadoEm.UTC(), AtualizadoEm: l.AtualizadoEm.UTC(),
		MotivoCancelamento: l.MotivoCancelamento,
	}
	if l.DevolucaoDeID != nil {
		p := parseUUID(*l.DevolucaoDeID)
		out.DevolucaoDeId = &p
	}
	if l.CanceladoPor != nil {
		p := parseUUID(*l.CanceladoPor)
		out.CanceladoPor = &p
	}
	if l.CanceladoEm != nil {
		t := l.CanceladoEm.UTC()
		out.CanceladoEm = &t
	}
	return out
}

func (h *Handler) ListLancamentos(ctx context.Context, _ ListLancamentosRequestObject) (ListLancamentosResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	lancamentos, err := h.M.ListarLancamentos.Execute(requestContext(ctx), finapp.ListarLancamentosInput{Actor: info.Principal})
	if err != nil {
		return nil, err
	}
	out := LancamentoList{Items: make([]Lancamento, len(lancamentos))}
	for i, l := range lancamentos {
		out.Items[i] = lancamentoOut(l)
	}
	return ListLancamentos200JSONResponse(out), nil
}

func (h *Handler) CreateLancamento(ctx context.Context, req CreateLancamentoRequestObject) (CreateLancamentoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	criado, err := h.M.CriarLancamento.Execute(requestContext(ctx), finapp.CriarLancamentoInput{
		Actor: info.Principal, Tipo: domain.TipoLancamento(req.Body.Tipo), ContaID: req.Body.ContaId.String(),
		ValorBrutoCents: int64(req.Body.ValorBrutoCents), TaxaCents: int64(req.Body.TaxaCents),
		FormaPagamento: domain.FormaPagamento(req.Body.FormaPagamento),
	})
	if err != nil {
		return nil, err
	}
	return CreateLancamento201JSONResponse(lancamentoOut(criado)), nil
}

func (h *Handler) UpdateLancamento(ctx context.Context, req UpdateLancamentoRequestObject) (UpdateLancamentoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	editado, err := h.M.EditarLancamento.Execute(requestContext(ctx), finapp.EditarLancamentoInput{
		Actor: info.Principal, ID: req.Id.String(), ContaID: req.Body.ContaId.String(),
		ValorBrutoCents: int64(req.Body.ValorBrutoCents), TaxaCents: int64(req.Body.TaxaCents),
		FormaPagamento: domain.FormaPagamento(req.Body.FormaPagamento),
	})
	if err != nil {
		return nil, err
	}
	return UpdateLancamento200JSONResponse(lancamentoOut(editado)), nil
}

func (h *Handler) CreateDevolucao(ctx context.Context, req CreateDevolucaoRequestObject) (CreateDevolucaoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	criado, err := h.M.CriarDevolucao.Execute(requestContext(ctx), finapp.CriarDevolucaoInput{
		Actor: info.Principal, ContaID: req.Body.ContaId.String(),
		ValorBrutoCents: int64(req.Body.ValorBrutoCents), TaxaCents: int64(req.Body.TaxaCents),
		FormaPagamento: domain.FormaPagamento(req.Body.FormaPagamento), DevolucaoDeID: req.Body.DevolucaoDeId.String(),
	})
	if err != nil {
		return nil, err
	}
	return CreateDevolucao201JSONResponse(lancamentoOut(criado)), nil
}

func (h *Handler) ReceiveLancamento(ctx context.Context, req ReceiveLancamentoRequestObject) (ReceiveLancamentoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := h.M.ReceberLancamento.Execute(requestContext(ctx), finapp.ReceberLancamentoInput{Actor: info.Principal, ID: req.Id.String()}); err != nil {
		return nil, err
	}
	return ReceiveLancamento204Response{}, nil
}

func (h *Handler) PayLancamento(ctx context.Context, req PayLancamentoRequestObject) (PayLancamentoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := h.M.PagarLancamento.Execute(requestContext(ctx), finapp.PagarLancamentoInput{Actor: info.Principal, ID: req.Id.String()}); err != nil {
		return nil, err
	}
	return PayLancamento204Response{}, nil
}

func (h *Handler) CancelLancamento(ctx context.Context, req CancelLancamentoRequestObject) (CancelLancamentoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	if _, err := h.M.CancelarLancamento.Execute(requestContext(ctx), finapp.CancelarLancamentoInput{
		Actor: info.Principal, ID: req.Id.String(), Motivo: req.Body.Reason,
	}); err != nil {
		return nil, err
	}
	return CancelLancamento204Response{}, nil
}
