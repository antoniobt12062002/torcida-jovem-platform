"use client";

import { useId, useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { Button } from "@/components/ui/button";
import { isApiError } from "@/lib/api/problem";

import { usersCatalog } from "./errors";
import { type UsersListFilters, useUsers } from "./hooks";
import { ALL_ROLES, type Role, roleLabel } from "./roles";
import { UsersTable } from "./users-table";

// Administração de usuários (USR-01): filtros enviados à API, lista paginada
// por cursor e "Carregar mais".

type ActiveFilter = "todos" | "ativos" | "inativos";

const ACTIVE_VALUES: Readonly<Record<ActiveFilter, boolean | undefined>> = {
  todos: undefined,
  ativos: true,
  inativos: false,
};

const SELECT_CLASS =
  "h-8 rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 dark:bg-input/30";

function toFilters(active: ActiveFilter, role: Role | ""): UsersListFilters {
  const filters: UsersListFilters = {};
  const activeValue = ACTIVE_VALUES[active];
  if (activeValue !== undefined) filters.active = activeValue;
  if (role) filters.role = role;
  return filters;
}

export function UsersPage() {
  const activeId = useId();
  const roleId = useId();
  const [active, setActive] = useState<ActiveFilter>("todos");
  const [role, setRole] = useState<Role | "">("");
  const query = useUsers(toFilters(active, role));
  const users = query.data?.pages.flatMap((page) => page.items) ?? [];

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-wrap items-end justify-between gap-4">
        <h1 className="text-xl font-semibold">Usuários</h1>
      </div>

      <div className="flex flex-wrap gap-4">
        <div className="flex flex-col gap-1">
          <label htmlFor={activeId} className="text-sm font-medium">
            Situação
          </label>
          <select
            id={activeId}
            className={SELECT_CLASS}
            value={active}
            onChange={(event) => setActive(event.target.value as ActiveFilter)}
          >
            <option value="todos">Todos</option>
            <option value="ativos">Ativos</option>
            <option value="inativos">Inativos</option>
          </select>
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor={roleId} className="text-sm font-medium">
            Papel
          </label>
          <select
            id={roleId}
            className={SELECT_CLASS}
            value={role}
            onChange={(event) => setRole(event.target.value as Role | "")}
          >
            <option value="">Todos</option>
            {ALL_ROLES.map((value) => (
              <option key={value} value={value}>
                {roleLabel(value)}
              </option>
            ))}
          </select>
        </div>
      </div>

      {query.isPending ? (
        <LoadingState />
      ) : query.isError && !query.data ? (
        <ErrorState error={query.error} catalog={usersCatalog} onRetry={() => void query.refetch()} />
      ) : users.length === 0 ? (
        <EmptyState title="Nenhum usuário encontrado." />
      ) : (
        <div className="flex flex-col gap-4">
          <UsersTable users={users} />
          {query.isFetchNextPageError && isApiError(query.error) ? (
            <ApiErrorAlert
              error={query.error}
              catalog={usersCatalog}
              onRetry={() => void query.fetchNextPage()}
            />
          ) : null}
          {query.hasNextPage ? (
            <Button
              variant="outline"
              className="self-center"
              disabled={query.isFetchingNextPage}
              onClick={() => void query.fetchNextPage()}
            >
              {query.isFetchingNextPage ? "Carregando…" : "Carregar mais"}
            </Button>
          ) : null}
        </div>
      )}
    </div>
  );
}
