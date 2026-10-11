import type { NavItem, NavSection } from "@/components/app/nav";

// Itens de navegação do módulo identity. A INT os liga em nav-items.ts.

export const USERS_PATH = "/admin/usuarios";

export const identityNav: NavItem[] = [
  { label: "Usuários", href: USERS_PATH, permission: "identity:user:read" },
];

/** Os mesmos itens agrupados sob "Administração". */
export const identityNavSection: NavSection = { label: "Administração", items: identityNav };
