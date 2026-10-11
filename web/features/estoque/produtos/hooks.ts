"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/estoque";
import type { ApiError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";

export type Produto = components["schemas"]["Produto"];
export type CreateProdutoRequest = components["schemas"]["CreateProdutoRequest"];

async function fetchProdutos(): Promise<Produto[]> {
  const data = await unwrap(api.estoque.GET("/api/v1/estoque/produtos"));
  return data.items;
}

/** GET /produtos: a lista inteira (WEB-D-016). */
export function useProdutos() {
  return useQuery<Produto[], ApiError>({
    queryKey: queryKeys.estoque.produtos,
    queryFn: fetchProdutos,
  });
}

/**
 * Produto lido do cache da lista (não há GET /produtos/{id}). `produto` é
 * null quando a lista chegou e o id não está nela.
 */
export function useProduto(id: string) {
  const query = useProdutos();
  const produto = query.data ? (query.data.find((p) => p.id === id) ?? null) : undefined;
  return { ...query, produto };
}

/** GET /produtos/{id}/saldo (EWB-01 AC4). */
export function useSaldo(produtoId: string) {
  return useQuery<number, ApiError>({
    queryKey: queryKeys.estoque.saldo(produtoId),
    queryFn: async () => {
      const data = await unwrap(
        api.estoque.GET("/api/v1/estoque/produtos/{id}/saldo", { params: { path: { id: produtoId } } }),
      );
      return data.saldo;
    },
  });
}

/** POST /produtos (EWB-01 AC2); a lista é rebuscada depois da criação. */
export function useCreateProduto() {
  const queryClient = useQueryClient();
  return useMutation<Produto, ApiError, CreateProdutoRequest>({
    mutationFn: (body) => unwrap(api.estoque.POST("/api/v1/estoque/produtos", { body })),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.estoque.produtos }),
  });
}
