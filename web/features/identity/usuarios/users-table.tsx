"use client";

import { activeStatus, StatusBadge } from "@/components/app/status-badge";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";

import type { User } from "./hooks";
import { roleLabel } from "./roles";

// Linhas da lista de usuários (USR-01 AC1).

export function UsersTable({ users }: { users: readonly User[] }) {
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
        </TableRow>
      </TableHeader>
      <TableBody>
        {users.map((user) => (
          <TableRow key={user.id}>
            <TableCell className="font-medium">{user.name}</TableCell>
            <TableCell>{user.email}</TableCell>
            <TableCell>
              <StatusBadge status={activeStatus(user.active)} />
            </TableCell>
            <TableCell className="whitespace-normal">{user.roles.map(roleLabel).join(", ")}</TableCell>
            <TableCell>{user.admin_membership ? "Sim" : "Não"}</TableCell>
            <TableCell>{user.must_change_password ? "Pendente" : "Não"}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
