"use client";

import { EllipsisIcon } from "lucide-react";

import { activeStatus, StatusBadge } from "@/components/app/status-badge";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import { ACTION_LABELS, type UserAction } from "./actions";
import type { User } from "./hooks";
import { roleLabel } from "./roles";

// Linhas da lista de usuários (USR-01 AC1) e menu de ações por linha.

const DESTRUCTIVE: ReadonlySet<UserAction> = new Set(["deactivate"]);

export type UsersTableProps = {
  users: readonly User[];
  /** Ações oferecidas para cada linha (vazio = sem menu). */
  actionsFor: (user: User) => readonly UserAction[];
  onAction: (action: UserAction, user: User) => void;
};

export function UsersTable({ users, actionsFor, onAction }: UsersTableProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Nome</TableHead>
          <TableHead>E-mail</TableHead>
          <TableHead>Situação</TableHead>
          <TableHead>Papéis</TableHead>
          <TableHead>Acesso administrativo</TableHead>
          <TableHead>Troca de senha</TableHead>
          <TableHead>
            <span className="sr-only">Ações</span>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {users.map((user) => {
          const actions = actionsFor(user);
          return (
            <TableRow key={user.id}>
              <TableCell className="font-medium">{user.name}</TableCell>
              <TableCell>{user.email}</TableCell>
              <TableCell>
                <StatusBadge status={activeStatus(user.active)} />
              </TableCell>
              <TableCell className="whitespace-normal">{user.roles.map(roleLabel).join(", ")}</TableCell>
              <TableCell>{user.admin_membership ? "Sim" : "Não"}</TableCell>
              <TableCell>{user.must_change_password ? "Pendente" : "Não"}</TableCell>
              <TableCell data-column="acoes" className="text-right">
                {actions.length > 0 ? (
                  <DropdownMenu>
                    <DropdownMenuTrigger
                      render={<Button variant="ghost" size="icon-sm" aria-label={`Ações de ${user.name}`} />}
                    >
                      <EllipsisIcon />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" className="w-auto">
                      {actions.map((action) => (
                        <DropdownMenuItem
                          key={action}
                          variant={DESTRUCTIVE.has(action) ? "destructive" : "default"}
                          onClick={() => onAction(action, user)}
                        >
                          {ACTION_LABELS[action]}
                        </DropdownMenuItem>
                      ))}
                    </DropdownMenuContent>
                  </DropdownMenu>
                ) : null}
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
