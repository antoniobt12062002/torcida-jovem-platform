import { estoqueNav } from "@/features/estoque/nav";
import { financeiroNav } from "@/features/financeiro/nav";
import { identityNavSection } from "@/features/identity/nav";
import { routes } from "@/lib/session/routes";

import type { NavEntry, NavItem } from "./nav";

// Navegação global (INT-01 AC1). Cada módulo define os próprios itens em
// features/<modulo>/nav.ts; aqui eles só são juntados, na ordem da spec.

export const inicioNavItem: NavItem = { label: "Início", href: routes.inicio };

export const navItems: readonly NavEntry[] = [inicioNavItem, financeiroNav, estoqueNav, identityNavSection];
