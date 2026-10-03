package financeirohttp

import (
	"context"

	finapp "github.com/antoniobt12062002/torcida-jovem-platform/api/internal/financeiro/app"
)

func (h *Handler) GetSaldo(ctx context.Context, _ GetSaldoRequestObject) (GetSaldoResponseObject, error) {
	info, err := sessionOf(ctx)
	if err != nil {
		return nil, err
	}
	saldo, err := h.M.ConsultarSaldo.Execute(requestContext(ctx), finapp.ConsultarSaldoInput{Actor: info.Principal})
	if err != nil {
		return nil, err
	}
	return GetSaldo200JSONResponse(Saldo{SaldoCents: saldo}), nil
}
