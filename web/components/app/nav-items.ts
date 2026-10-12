import { House } from "lucide-react";

import { estoqueNav } from "@/features/estoque/nav";
import { financeiroNav } from "@/features/financeiro/nav";
import { identityNavSection } from "@/features/identity/nav";
import { routes } from "@/lib/session/routes";

import type { NavEntry, NavItem } from "./nav";

// Navegação global (INT-01 AC1). Cada módulo define os próprios itens em
// features/<modulo>/nav.ts, com ícone e descrição; aqui eles só são juntados,
// na ordem da spec.

export const inicioNavItem: NavItem = {
  label: "Início",
  href: routes.inicio,
  icon: House,
  description: "Atalhos para as áreas que a sua conta pode usar.",
};

export const navItems: readonly NavEntry[] = [inicioNavItem, financeiroNav, estoqueNav, identityNavSection];
