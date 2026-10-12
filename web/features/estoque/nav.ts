import { PackageIcon } from "lucide-react";

import type { NavSection } from "@/components/app/nav";

// Itens de navegação do estoque. A INT os liga à navegação global.

export const estoqueNav: NavSection = {
  label: "Estoque",
  items: [
    {
      label: "Produtos",
      href: "/estoque/produtos",
      permission: "estoque:produto:read",
      icon: PackageIcon,
      description: "Produtos, saldos e movimentações do estoque.",
    },
  ],
};
