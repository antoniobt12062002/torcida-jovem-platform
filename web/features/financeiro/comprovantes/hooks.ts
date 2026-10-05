"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { api, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/financeiro";
import { queryKeys } from "@/lib/api/query-keys";

// Comprovantes de um lançamento (FWB-05). A URL assinada de download é buscada
// no clique e nunca entra no cache de consultas (expira).

export type Comprovante = components["schemas"]["Comprovante"];

export function useComprovantes(lancamentoId: string) {
  return useQuery({
    queryKey: queryKeys.financeiro.comprovantes(lancamentoId),
    queryFn: () =>
      unwrap(
        api.financeiro.GET("/api/v1/financeiro/lancamentos/{id}/comprovantes", {
          params: { path: { id: lancamentoId } },
        }),
      ),
  });
}

/**
 * Envia o arquivo como multipart/form-data no campo `file`. O navegador define
 * o Content-Type com o boundary; o CSRF vem do middleware do cliente.
 */
export function useUploadComprovante(lancamentoId: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) =>
      unwrap(
        api.financeiro.POST("/api/v1/financeiro/lancamentos/{id}/comprovantes", {
          params: { path: { id: lancamentoId } },
          // O contrato descreve o campo binário como string; o corpo real é o FormData abaixo.
          body: { file: file as unknown as string },
          bodySerializer: () => {
            const form = new FormData();
            form.append("file", file, file.name);
            return form;
          },
        }),
      ),
    onSuccess: () =>
      queryClient.invalidateQueries({ queryKey: queryKeys.financeiro.comprovantes(lancamentoId) }),
  });
}

/** Busca a URL assinada no momento do pedido; sem cache. */
export async function downloadComprovante(documentId: string): Promise<string> {
  const { url } = await unwrap(
    api.financeiro.GET("/api/v1/financeiro/comprovantes/{documentId}/url", {
      params: { path: { documentId } },
    }),
  );
  return url;
}
