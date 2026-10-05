"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { useMemo } from "react";

import { shortId } from "@/components/app/format";
import type { components } from "@/lib/api/financeiro";
import { api, unwrap } from "@/lib/api/client";
import { isApiError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";
import { usePermission } from "@/lib/session";

import { STALE_CODES } from "./errors";
import type { Conta, ContaList, Lancamento, LancamentoList } from "./labels";

// Consultas e escritas de lançamentos. Contas e saldo passam só pelas chaves
// compartilhadas de lib/api/query-keys.ts (WEB-D-010).

export type CreateLancamentoBody = components["schemas"]["CreateLancamentoRequest"];
export type UpdateLancamentoBody = components["schemas"]["UpdateLancamentoRequest"];

function fetchLancamentos(): Promise<LancamentoList> {
  return unwrap(api.financeiro.GET("/api/v1/financeiro/lancamentos"));
}

/** Mesma chave e mesma forma de retorno (`ContaList`) da tela de contas. */
function fetchContas(): Promise<ContaList> {
  return unwrap(api.financeiro.GET("/api/v1/financeiro/contas"));
}

export function useLancamentos() {
  return useQuery({ queryKey: queryKeys.financeiro.lancamentos, queryFn: fetchLancamentos });
}

/** Um lançamento lido do cache da lista (não há GET por id); `null` se não estiver nela. */
export function useLancamento(id: string) {
  return useQuery({
    queryKey: queryKeys.financeiro.lancamentos,
    queryFn: fetchLancamentos,
    select: (data: LancamentoList): Lancamento | null => data.items.find((l) => l.id === id) ?? null,
  });
}

/** Contas para nomes e opções; só consultadas com `financeiro:conta:read`. */
export function useContasOptions() {
  const canRead = usePermission("financeiro:conta:read");
  return useQuery({ queryKey: queryKeys.financeiro.contas, queryFn: fetchContas, enabled: canRead });
}

/** Nome da conta pelo id; sem a conta carregada, o identificador curto. */
export function useContaNome(): (id: string) => string {
  const contas = useContasOptions();
  const byId = useMemo(
    () => new Map<string, Conta>((contas.data?.items ?? []).map((c) => [c.id, c])),
    [contas.data],
  );
  return (id) => byId.get(id)?.nome ?? shortId(id);
}

function useInvalidateOnStale() {
  const queryClient = useQueryClient();
  return (error: unknown) => {
    if (isApiError(error) && STALE_CODES.has(error.code)) {
      void queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.lancamentos });
    }
  };
}

export function useCreateLancamento() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (body: CreateLancamentoBody) =>
      unwrap(api.financeiro.POST("/api/v1/financeiro/lancamentos", { body })),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.lancamentos }),
  });
}

/** Depois de uma mudança de status: lista e saldo são buscados de novo. */
function useInvalidateStatusChange() {
  const queryClient = useQueryClient();
  return () =>
    Promise.all([
      queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.lancamentos }),
      queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.saldo }),
    ]);
}

export function useReceberLancamento(id: string) {
  const onSuccess = useInvalidateStatusChange();
  const onStale = useInvalidateOnStale();
  return useMutation({
    mutationFn: () =>
      unwrap(api.financeiro.POST("/api/v1/financeiro/lancamentos/{id}/receive", { params: { path: { id } } })),
    onSuccess,
    onError: onStale,
  });
}

export function usePagarLancamento(id: string) {
  const onSuccess = useInvalidateStatusChange();
  const onStale = useInvalidateOnStale();
  return useMutation({
    mutationFn: () =>
      unwrap(api.financeiro.POST("/api/v1/financeiro/lancamentos/{id}/pay", { params: { path: { id } } })),
    onSuccess,
    onError: onStale,
  });
}

export function useCancelarLancamento(id: string) {
  const onSuccess = useInvalidateStatusChange();
  const onStale = useInvalidateOnStale();
  return useMutation({
    mutationFn: (reason: string) =>
      unwrap(
        api.financeiro.POST("/api/v1/financeiro/lancamentos/{id}/cancel", {
          params: { path: { id } },
          body: { reason },
        }),
      ),
    onSuccess,
    onError: onStale,
  });
}

export type CreateDevolucaoBody = components["schemas"]["CreateDevolucaoRequest"];

export function useRegistrarDevolucao() {
  const queryClient = useQueryClient();
  const onStale = useInvalidateOnStale();
  return useMutation({
    mutationFn: (body: CreateDevolucaoBody) =>
      unwrap(api.financeiro.POST("/api/v1/financeiro/lancamentos/devolucoes", { body })),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.lancamentos }),
    onError: onStale,
  });
}

export function useUpdateLancamento(id: string) {
  const queryClient = useQueryClient();
  const onStale = useInvalidateOnStale();
  return useMutation({
    mutationFn: (body: UpdateLancamentoBody) =>
      unwrap(
        api.financeiro.PUT("/api/v1/financeiro/lancamentos/{id}", { params: { path: { id } }, body }),
      ),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.lancamentos }),
    onError: onStale,
  });
}
