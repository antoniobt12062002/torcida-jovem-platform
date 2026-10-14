"use client";

import { PlusIcon } from "lucide-react";
import { useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { FilterBar } from "@/components/app/filter-bar";
import { PageHeader } from "@/components/app/page-header";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { Button } from "@/components/ui/button";
import { Field, FieldLabel } from "@/components/ui/field";
import { NativeSelect } from "@/components/ui/native-select";
import { ShowMore } from "@/components/ui/show-more";
import { isApiError } from "@/lib/api/problem";
import { useCan, useSession } from "@/lib/session/session-provider";
import { codecs, useQueryState } from "@/lib/url-state";

import { availableActions, type UserAction } from "./actions";
import { GrantAdminDialog, RevokeAdminDialog, SetRolesDialog } from "./dialogs/admin-dialogs";
import { CreateUserDialog } from "./dialogs/create-user-dialog";
import { DeactivateUserDialog, ReactivateUserDialog } from "./dialogs/status-dialogs";
import { TemporaryPasswordDialog } from "./dialogs/temporary-password-dialog";
import { usersCatalog } from "./errors";
import { type User, type UsersListFilters, useUsers } from "./hooks";
import { ALL_ROLES, type Role, roleLabel } from "./roles";
import { UsersTable } from "./users-table";

// Administração de usuários (USR-01 a USR-05; DSM-02). Os filtros ficam na URL
// (`situacao` e `papel`, DSP-05 AC2 a AC4) e seguem para a API como `active` e
// `role`; a lista é paginada por cursor ("Mostrar mais", DSP-05 AC14) e as
// ações de cada linha abrem diálogos. Quem usa fica dentro de `<Suspense>`.

type OpenAction = { action: UserAction; user: User };

function ActionDialog({ open, onClose }: { open: OpenAction; onClose: () => void }) {
  const props = {
    user: open.user,
    open: true,
    onOpenChange: (next: boolean) => {
      if (!next) onClose();
    },
  };
  switch (open.action) {
    case "deactivate":
      return <DeactivateUserDialog {...props} />;
    case "reactivate":
      return <ReactivateUserDialog {...props} />;
    case "grant":
      return <GrantAdminDialog {...props} />;
    case "setRoles":
      return <SetRolesDialog {...props} />;
    case "revoke":
      return <RevokeAdminDialog {...props} />;
    case "resetPassword":
      return <TemporaryPasswordDialog {...props} />;
  }
}

const SITUACOES = ["ativos", "inativos"] as const;
type Situacao = (typeof SITUACOES)[number];

const SITUACAO_LABELS: Readonly<Record<Situacao, string>> = { ativos: "Ativos", inativos: "Inativos" };
const ACTIVE_OF: Readonly<Record<Situacao, boolean>> = { ativos: true, inativos: false };

const situacaoCodec = codecs.oneOf(SITUACOES, "");
const papelCodec = codecs.oneOf(ALL_ROLES, "");

function toFilters(situacao: Situacao | "", papel: Role | ""): UsersListFilters {
  const filters: UsersListFilters = {};
  if (situacao) filters.active = ACTIVE_OF[situacao];
  if (papel) filters.role = papel;
  return filters;
}

export function UsersPage() {
  const [situacao, setSituacao] = useQueryState("situacao", situacaoCodec);
  const [papel, setPapel] = useQueryState("papel", papelCodec);
  const query = useUsers(toFilters(situacao, papel));
  const users = query.data?.pages.flatMap((page) => page.items) ?? [];
  const { context } = useSession();
  const sessionUserId = context?.user.id;
  const permissions = context?.permissions ?? [];
  const canCreate = useCan("identity:user:create");
  const [creating, setCreating] = useState(false);
  const [openAction, setOpenAction] = useState<OpenAction | null>(null);

  const activeCount = [situacao, papel].filter((value) => value !== "").length;

  function clearFilters() {
    setSituacao("");
    setPapel("");
  }

  function content() {
    if (query.isPending) return <LoadingState shape="table" />;
    if (query.isError && !query.data) {
      return <ErrorState error={query.error} catalog={usersCatalog} onRetry={() => void query.refetch()} />;
    }
    if (users.length === 0) {
      return activeCount > 0 ? (
        <EmptyState kind="no-results" description="Mude a situação ou o papel." onClear={clearFilters} />
      ) : (
        <EmptyState
          title="Nenhum usuário encontrado."
          description="As pessoas cadastradas aparecem aqui, com a situação e os papéis de cada uma."
        />
      );
    }
    return (
      <>
        <UsersTable
          users={users}
          actionsFor={(user) => availableActions(user, sessionUserId, permissions)}
          onAction={(action, user) => setOpenAction({ action, user })}
        />
        {query.isFetchNextPageError && isApiError(query.error) ? (
          <ApiErrorAlert error={query.error} catalog={usersCatalog} onRetry={() => void query.fetchNextPage()} />
        ) : null}
        {query.hasNextPage ? (
          <ShowMore
            shown={users.length}
            noun={["usuário", "usuários"]}
            loading={query.isFetchingNextPage}
            onShowMore={() => void query.fetchNextPage()}
          />
        ) : null}
      </>
    );
  }

  return (
    <>
      <PageHeader
        title="Usuários"
        description="Contas de acesso, situação, papéis e acesso administrativo."
        actions={
          canCreate ? (
            <Button onClick={() => setCreating(true)}>
              <PlusIcon aria-hidden="true" />
              Novo usuário
            </Button>
          ) : undefined
        }
      />
      {canCreate ? <CreateUserDialog open={creating} onOpenChange={setCreating} /> : null}

      <div className="flex flex-col gap-4">
        <FilterBar
          activeCount={activeCount}
          resultCount={users.length}
          resultNoun={["usuário", "usuários"]}
          onClear={clearFilters}
        >
          <Field>
            <FieldLabel>Situação</FieldLabel>
            <NativeSelect
              name="situacao"
              autoComplete="off"
              value={situacao}
              onChange={(event) => setSituacao(situacaoCodec.parse(event.target.value))}
            >
              <option value="">Todos</option>
              {SITUACOES.map((value) => (
                <option key={value} value={value}>
                  {SITUACAO_LABELS[value]}
                </option>
              ))}
            </NativeSelect>
          </Field>
          <Field>
            <FieldLabel>Papel</FieldLabel>
            <NativeSelect
              name="papel"
              autoComplete="off"
              value={papel}
              onChange={(event) => setPapel(papelCodec.parse(event.target.value))}
            >
              <option value="">Todos</option>
              {ALL_ROLES.map((value) => (
                <option key={value} value={value}>
                  {roleLabel(value)}
                </option>
              ))}
            </NativeSelect>
          </Field>
        </FilterBar>
        {content()}
      </div>

      {openAction ? (
        <ActionDialog
          key={`${openAction.action}:${openAction.user.id}`}
          open={openAction}
          onClose={() => setOpenAction(null)}
        />
      ) : null}
    </>
  );
}
