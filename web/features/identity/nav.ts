import { UsersIcon } from "lucide-react";

import type { NavItem, NavSection } from "@/components/app/nav";

// Itens de navegação do módulo identity. A INT os liga em nav-items.ts. O
// ícone e a descrição servem ao shell e ao atalho da página inicial (DSS-01).

export const USERS_PATH = "/admin/usuarios";

export const USERS_DESCRIPTION = "Contas de acesso, papéis e acesso administrativo.";

export const identityNav: NavItem[] = [
  {
    label: "Usuários",
    href: USERS_PATH,
    permission: "identity:user:read",
    icon: UsersIcon,
    description: USERS_DESCRIPTION,
  },
];

/** Os mesmos itens agrupados sob "Administração". */
export const identityNavSection: NavSection = { label: "Administração", items: identityNav };
