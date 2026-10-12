"use client";

import {
  KeyRoundIcon,
  type LucideIcon,
  ShieldMinusIcon,
  ShieldPlusIcon,
  UserCheckIcon,
  UserCogIcon,
  UserXIcon,
} from "lucide-react";

import { type Column, DataList } from "@/components/app/data-list";
import { RoleBadge } from "@/components/app/role-badge";
import type { RowAction } from "@/components/app/row-actions";
import { activeStatusDef, StatusBadge } from "@/components/app/status-badge";

import { ACTION_LABELS, type UserAction } from "./actions";
import type { User } from "./hooks";
import { roleLabel } from "./roles";

// Listagem de usuários (USR-01 AC1; DSM-02 AC1 e AC3; DSP-05 AC5 a AC7 e
// AC10). Colunas por prioridade (design/05-migracao-web-v1.md): nome e
// situação sempre; e-mail e papéis a partir de 768 px; acesso administrativo e
// troca de senha só na tabela completa. Na lista empilhada: título nome,
// campos e-mail e papéis, situação como status. Todas as ações ficam no menu
// "Ações de {nome}"; a linha da própria pessoa não tem menu.

const ACTION_ICONS: Readonly<Record<UserAction, LucideIcon>> = {
  grant: ShieldPlusIcon,
  setRoles: UserCogIcon,
  revoke: ShieldMinusIcon,
  resetPassword: KeyRoundIcon,
  deactivate: UserXIcon,
  reactivate: UserCheckIcon,
};

// Desativar tira da pessoa o acesso ao sistema: vai por último, separado. As
// demais mantêm a ordem de `availableActions`.
const DESTRUCTIVE: ReadonlySet<UserAction> = new Set(["deactivate"]);

function RoleList({ user }: { user: User }) {
  return (
    <span className="inline-flex flex-wrap gap-1">
      {user.roles.map((role, index) => (
        <span key={role} className="inline-flex">
          {/* O separador mantém a leitura "Associado, Tesouraria" nos leitores de tela. */}
          {index > 0 ? <span className="sr-only">, </span> : null}
          <RoleBadge label={roleLabel(role)} />
        </span>
      ))}
    </span>
  );
}

export const USER_COLUMNS: readonly Column<User>[] = [
  { id: "nome", header: "Nome", priority: 1, inList: "title", cell: (user) => user.name },
  { id: "email", header: "E-mail", priority: 2, inList: "field", cell: (user) => user.email },
  {
    id: "situacao",
    header: "Situação",
    priority: 1,
    inList: "meta",
    cell: (user) => <StatusBadge status={activeStatusDef(user.active)} />,
  },
  { id: "papeis", header: "Papéis", priority: 2, inList: "field", cell: (user) => <RoleList user={user} /> },
  {
    id: "admin",
    header: "Acesso administrativo",
    priority: 3,
    inList: false,
    wrap: "nowrap",
    cell: (user) => (user.admin_membership ? "Sim" : "Não"),
  },
  {
    id: "troca",
    header: "Troca de senha",
    priority: 3,
    inList: false,
    wrap: "nowrap",
    cell: (user) => (user.must_change_password ? "Pendente" : "Não"),
  },
];

export type UsersTableProps = {
  users: readonly User[];
  /** Ações oferecidas para cada linha (vazio = sem menu). */
  actionsFor: (user: User) => readonly UserAction[];
  onAction: (action: UserAction, user: User) => void;
};

export function UsersTable({ users, actionsFor, onAction }: UsersTableProps) {
  return (
    <DataList
      label="Usuários"
      columns={USER_COLUMNS}
      rows={users}
      rowKey={(user) => user.id}
      rowLabel={(user) => user.name}
      moreActions={(user) =>
        actionsFor(user).map(
          (action): RowAction => ({
            label: ACTION_LABELS[action],
            icon: ACTION_ICONS[action],
            destructive: DESTRUCTIVE.has(action),
            onSelect: () => onAction(action, user),
          }),
        )
      }
    />
  );
}
