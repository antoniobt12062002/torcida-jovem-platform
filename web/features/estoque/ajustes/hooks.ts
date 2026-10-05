"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/estoque";
import type { ApiError } from "@/lib/api/problem";

import { invalidateEstoqueDoProduto, type Movimentacao } from "../movimentacoes/hooks";

export type CreateAjusteRequest = components["schemas"]["CreateAjusteRequest"];

/**
 * POST /ajustes (EWB-03 AC1). A API registra o ajuste com AJUSTE_MANUAL; o
 * histórico e o saldo do produto são invalidados depois.
 */
export function useAjustarEstoque() {
  const queryClient = useQueryClient();
  return useMutation<Movimentacao, ApiError, CreateAjusteRequest>({
    mutationFn: (body) => unwrap(api.estoque.POST("/api/v1/estoque/ajustes", { body })),
    onSuccess: (_data, body) => invalidateEstoqueDoProduto(queryClient, body.produto_id),
  });
}
