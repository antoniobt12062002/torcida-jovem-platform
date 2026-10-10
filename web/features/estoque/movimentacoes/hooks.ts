"use client";

import { type QueryClient, useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/estoque";
import type { ApiError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";

import { type MovimentacaoManual, movimentacaoPayload } from "./payload";

export type Movimentacao = components["schemas"]["Movimentacao"];

/** GET /produtos/{id}/movimentacoes, na ordem devolvida pela API (EWB-02 AC1). */
export function useMovimentacoes(produtoId: string) {
  return useQuery<Movimentacao[], ApiError>({
    queryKey: queryKeys.estoque.movimentacoes(produtoId),
    queryFn: async () => {
      const data = await unwrap(
        api.estoque.GET("/api/v1/estoque/produtos/{id}/movimentacoes", {
          params: { path: { id: produtoId } },
        }),
      );
      return data.items;
    },
  });
}

/** Histórico e saldo do produto ficam desatualizados depois de qualquer registro (EWB-02 AC6). */
export function invalidateEstoqueDoProduto(queryClient: QueryClient, produtoId: string) {
  return Promise.all([
    queryClient.invalidateQueries({ queryKey: queryKeys.estoque.movimentacoes(produtoId) }),
    queryClient.invalidateQueries({ queryKey: queryKeys.estoque.saldo(produtoId) }),
  ]);
}

/** POST /movimentacoes com o corpo de payload.ts (origem INVENTARIO). */
export function useRegistrarMovimentacao() {
  const queryClient = useQueryClient();
  return useMutation<Movimentacao, ApiError, MovimentacaoManual>({
    mutationFn: (input) =>
      unwrap(api.estoque.POST("/api/v1/estoque/movimentacoes", { body: movimentacaoPayload(input) })),
    onSuccess: (_data, input) => invalidateEstoqueDoProduto(queryClient, input.produtoId),
  });
}
