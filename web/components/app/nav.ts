// Navegação da aplicação (FND-05 AC1, FND-04 AC3). A lista real de itens é
// montada pela INT em nav-items.ts; aqui ficam só o formato e o filtro.

import type { LucideIcon } from "lucide-react";

export type NavItem = {
  label: string;
  href: string;
  /** Permissão efetiva exigida. Sem ela, o item vale para qualquer pessoa autenticada. */
  permission?: string;
  /** Ícone do item na navegação do shell (DSS-01 AC3), definido por cada módulo. */
  icon?: LucideIcon;
  /** Frase que descreve a área no atalho da página inicial. */
  description?: string;
};

/** Grupo de itens com um título (por exemplo "Financeiro"). */
export type NavSection = {
  label: string;
  items: NavItem[];
};

export type NavEntry = NavItem | NavSection;

export function isNavSection(entry: NavEntry): entry is NavSection {
  return "items" in entry;
}

function allowed(item: NavItem, permissions: readonly string[]): boolean {
  return item.permission === undefined || permissions.includes(item.permission);
}

/**
 * Itens que a pessoa pode ver, pelas permissões efetivas (nunca pelo papel).
 * Seções sem nenhum item visível são removidas.
 */
export function visibleItems(entries: readonly NavEntry[], permissions: readonly string[]): NavEntry[] {
  const result: NavEntry[] = [];
  for (const entry of entries) {
    if (isNavSection(entry)) {
      const items = entry.items.filter((item) => allowed(item, permissions));
      if (items.length > 0) result.push({ ...entry, items });
    } else if (allowed(entry, permissions)) {
      result.push(entry);
    }
  }
  return result;
}
