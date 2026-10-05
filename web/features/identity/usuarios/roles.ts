import type { components } from "@/lib/api/identity";

// Papéis do contrato (enum Role) com rótulos em português.

export type Role = components["schemas"]["Role"];

export const ROLE_LABELS: Readonly<Record<Role, string>> = {
  ASSOCIADO: "Associado",
  PRESIDENTE: "Presidente",
  DIRETORIA: "Diretoria",
  TESOURARIA: "Tesouraria",
  ESTOQUE_LOJA: "Estoque e loja",
  EVENTOS: "Eventos",
  CONSELHO_FISCAL: "Conselho Fiscal",
  ADMIN_SISTEMA: "Administração do sistema",
};

/** Todos os papéis, na ordem do contrato (filtro da lista). */
export const ALL_ROLES: readonly Role[] = [
  "ASSOCIADO",
  "PRESIDENTE",
  "DIRETORIA",
  "TESOURARIA",
  "ESTOQUE_LOJA",
  "EVENTOS",
  "CONSELHO_FISCAL",
  "ADMIN_SISTEMA",
];

/**
 * Papéis administrativos oferecidos na promoção e na edição: todos menos
 * ASSOCIADO, que a API sempre mantém. A API recusa o que o ator não pode
 * conceder (privilege_escalation).
 */
export const ADMIN_ROLES: readonly Role[] = ALL_ROLES.filter((role) => role !== "ASSOCIADO");

export function roleLabel(role: Role): string {
  return ROLE_LABELS[role];
}
