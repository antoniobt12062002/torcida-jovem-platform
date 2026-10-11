"use client";

import { useMutation, useQueryClient } from "@tanstack/react-query";

import { api } from "@/lib/api/client";
import type { components } from "@/lib/api/identity";
import { networkError, toApiError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";

import { PasswordApiError } from "./errors";

type ChangePasswordRequest = components["schemas"]["ChangePasswordRequest"];

async function changePassword(body: ChangePasswordRequest): Promise<void> {
  let result;
  try {
    result = await api.identity.POST("/api/v1/auth/password", { body });
  } catch (cause) {
    throw new PasswordApiError(networkError(cause), null);
  }
  if (!result.response.ok) {
    throw new PasswordApiError(
      toApiError(result.response, result.error),
      result.response.headers.get("Retry-After"),
    );
  }
}

/**
 * POST /auth/password (ACS-01). Em sucesso, rebusca o contexto da sessão
 * (queryKeys.session.me): a troca desliga must_change_password. A mutação só
 * termina depois do contexto novo chegar.
 */
export function useChangePassword() {
  const queryClient = useQueryClient();
  return useMutation<void, PasswordApiError, ChangePasswordRequest>({
    mutationFn: changePassword,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.session.me }),
  });
}
