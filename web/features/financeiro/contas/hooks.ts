"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/financeiro";
import type { ApiError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";

// Consulta e escritas do plano de contas (FWB-01). A consulta guarda a
// ContaList como a API devolve, sob queryKeys.financeiro.contas: a frente de
// lançamentos lê a mesma chave com a mesma função de busca (WEB-D-010).

export type Conta = components["schemas"]["Conta"];
export type ContaList = components["schemas"]["ContaList"];
export type CreateContaRequest = components["schemas"]["CreateContaRequest"];

const CONTAS = "/api/v1/financeiro/contas";

function fetchContas(): Promise<ContaList> {
  return unwrap(api.financeiro.GET(CONTAS));
}

export function useContas() {
  return useQuery<ContaList, ApiError>({
    queryKey: queryKeys.financeiro.contas,
    queryFn: fetchContas,
  });
}

/** Escritas terminam só depois de a lista de contas ser buscada de novo. */
function useInvalidateContas() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.contas });
}

export function useCreateConta() {
  const invalidate = useInvalidateContas();
  return useMutation<Conta, ApiError, CreateContaRequest>({
    mutationFn: (body) => unwrap(api.financeiro.POST(CONTAS, { body })),
    onSuccess: invalidate,
  });
}

export function useRenameConta() {
  const invalidate = useInvalidateContas();
  return useMutation<Conta, ApiError, { id: string; nome: string }>({
    mutationFn: ({ id, nome }) =>
      unwrap(api.financeiro.PATCH("/api/v1/financeiro/contas/{id}", { params: { path: { id } }, body: { nome } })),
    onSuccess: invalidate,
  });
}

export function useDeactivateConta() {
  const invalidate = useInvalidateContas();
  return useMutation<void, ApiError, string>({
    mutationFn: async (id) => {
      await unwrap(api.financeiro.POST("/api/v1/financeiro/contas/{id}/deactivate", { params: { path: { id } } }));
    },
    onSuccess: invalidate,
  });
}
