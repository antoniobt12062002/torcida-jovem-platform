"use client";

import { useQuery } from "@tanstack/react-query";

import { api, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/financeiro";
import type { ApiError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";

// Saldo atual (FWB-02). As escritas de lançamentos invalidam esta chave.

export type Saldo = components["schemas"]["Saldo"];

export function useSaldo() {
  return useQuery<Saldo, ApiError>({
    queryKey: queryKeys.financeiro.saldo,
    queryFn: () => unwrap(api.financeiro.GET("/api/v1/financeiro/saldo")),
  });
}
