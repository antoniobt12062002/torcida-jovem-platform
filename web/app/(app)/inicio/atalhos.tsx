"use client";

import Link from "next/link";

import { isNavSection, type NavItem, type NavSection, visibleItems } from "@/components/app/nav";
import { navItems } from "@/components/app/nav-items";
import { LoadingState } from "@/components/app/states";
import { useSession } from "@/lib/session/session-provider";

// Um atalho por área visível na navegação, agrupado pelo módulo (INT-01 AC4).
export function Atalhos() {
  const { context } = useSession();
  if (!context) return <LoadingState />;

  const secoes = visibleItems(navItems, context.permissions).filter(isNavSection);

  if (secoes.length === 0) {
    return (
      <div className="flex flex-col gap-1 rounded-lg border border-dashed px-4 py-10 text-center">
        <p className="font-medium">Sua conta ainda não tem acesso a nenhum módulo.</p>
        <p className="text-sm text-muted-foreground">
          Se você precisa usar alguma área do sistema, fale com a administração.
        </p>
      </div>
    );
  }

  return (
    <nav aria-label="Atalhos das áreas" className="grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
      {secoes.map((secao) => (
        <Secao key={secao.label} secao={secao} />
      ))}
    </nav>
  );
}

function Secao({ secao }: { secao: NavSection }) {
  return (
    <section className="flex flex-col gap-2">
      <h2 className="text-sm font-medium text-muted-foreground uppercase">{secao.label}</h2>
      <ul className="flex flex-col gap-2">
        {secao.items.map((item) => (
          <li key={item.href}>
            <Atalho item={item} />
          </li>
        ))}
      </ul>
    </section>
  );
}

function Atalho({ item }: { item: NavItem }) {
  return (
    <Link href={item.href} className="block rounded-lg border px-4 py-3 text-sm font-medium hover:bg-muted">
      {item.label}
    </Link>
  );
}
