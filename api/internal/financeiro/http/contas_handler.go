package financeirohttp

import (
	"context"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"

	finapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
	"github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/domain"
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

func contaOut(c domain.Conta) Conta {
	out := Conta{Id: parseUUID(c.ID), Nome: c.Nome, Tipo: TipoConta(c.Tipo), Ativo: c.Ativo, CreatedAt: c.CreatedAt.UTC()}
	if c.ParentID != nil {
		p := parseUUID(*c.ParentID)
		out.ParentId = &p
	}
	return out
}

func (h *Handler) ListContas(ctx context.Context, _ ListContasRequestObject) (ListContasResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	contas, err := h.M.ListarContas.Execute(requestContext(ctx), finapp.ListarContasInput{Actor: info.Principal})
	if err != nil {
		return nil, err
	}
	out := ContaList{Items: make([]Conta, len(contas))}
	for i, c := range contas {
		out.Items[i] = contaOut(c)
	}
	return ListContas200JSONResponse(out), nil
}

func (h *Handler) CreateConta(ctx context.Context, req CreateContaRequestObject) (CreateContaResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	criada, err := h.M.CriarConta.Execute(requestContext(ctx), finapp.CriarContaInput{
		Actor: info.Principal, Tipo: domain.TipoConta(req.Body.Tipo), Nome: req.Body.Nome, ParentID: uuidPtrToString(req.Body.ParentId),
	})
	if err != nil {
		return nil, err
	}
	return CreateConta201JSONResponse(contaOut(criada)), nil
}

func (h *Handler) RenameConta(ctx context.Context, req RenameContaRequestObject) (RenameContaResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	renomeada, err := h.M.RenomearConta.Execute(requestContext(ctx), finapp.RenomearContaInput{
		Actor: info.Principal, ID: req.Id.String(), Nome: req.Body.Nome,
	})
	if err != nil {
		return nil, err
	}
	return RenameConta200JSONResponse(contaOut(renomeada)), nil
}

func (h *Handler) DeactivateConta(ctx context.Context, req DeactivateContaRequestObject) (DeactivateContaResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.M.DesativarConta.Execute(requestContext(ctx), finapp.DesativarContaInput{Actor: info.Principal, ID: req.Id.String()}); err != nil {
		return nil, err
	}
	return DeactivateConta204Response{}, nil
}
