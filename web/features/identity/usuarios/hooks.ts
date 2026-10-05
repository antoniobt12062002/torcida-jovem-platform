"use client";

import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";

import { api, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/identity";
import { queryKeys } from "@/lib/api/query-keys";

import type { Role } from "./roles";

export type User = components["schemas"]["User"];

/** Filtros da tela, enviados à API (USR-01 AC2). */
export type UsersListFilters = { active?: boolean; role?: Role };

/** Tamanho da página pedido à API (o padrão do contrato). */
export const PAGE_SIZE = 50;

/**
 * GET /users com cursor (USR-01). A chave é queryKeys.identity.users(filtros):
 * trocar um filtro troca a chave e a lista recomeça da primeira página.
 */
export function useUsers(filters: UsersListFilters) {
  return useInfiniteQuery({
    queryKey: queryKeys.identity.users(filters),
    initialPageParam: undefined as string | undefined,
    queryFn: ({ pageParam, signal }) =>
      unwrap(
        api.identity.GET("/api/v1/users", {
          params: {
            query: {
              limit: PAGE_SIZE,
              active: filters.active,
              role: filters.role,
              cursor: pageParam,
            },
          },
          signal,
        }),
      ),
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
  });
}

export type CreateUserRequest = components["schemas"]["CreateUserRequest"];

/**
 * Invalida todas as listas de usuários (qualquer filtro), para que reflitam a
 * mudança. A mutação só termina depois da nova consulta.
 */
function useInvalidateUsers() {
  const queryClient = useQueryClient();
  return () => queryClient.invalidateQueries({ queryKey: queryKeys.identity.usersAll });
}

/**
 * POST /users (USR-02). As variáveis da mutação trazem a senha inicial:
 * gcTime 0 tira a mutação do cache assim que o diálogo deixa de observá-la.
 */
export function useCreateUser() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: (body: CreateUserRequest) => unwrap(api.identity.POST("/api/v1/users", { body })),
    onSuccess: invalidate,
    gcTime: 0,
  });
}

/** POST /users/{id}/deactivate (USR-03 AC1). */
export function useDeactivateUser() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(api.identity.POST("/api/v1/users/{id}/deactivate", { params: { path: { id } } }));
    },
    onSuccess: invalidate,
  });
}

/** POST /users/{id}/reactivate (USR-03 AC2). */
export function useReactivateUser() {
  const invalidate = useInvalidateUsers();
  return useMutation({
    mutationFn: async (id: string) => {
      await unwrap(api.identity.POST("/api/v1/users/{id}/reactivate", { params: { path: { id } } }));
    },
    onSuccess: invalidate,
  });
}
