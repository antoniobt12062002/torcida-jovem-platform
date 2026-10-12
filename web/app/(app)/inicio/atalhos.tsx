"use client";

import Link from "next/link";

import { isNavSection, type NavItem, type NavSection, visibleItems } from "@/components/app/nav";
import { navItems } from "@/components/app/nav-items";
import { PageHeader } from "@/components/app/page-header";
import { EmptyState, LoadingState } from "@/components/app/states";
import { Card } from "@/components/ui/card";
import { SectionTitle } from "@/components/ui/typography";
import { useSession } from "@/lib/session/session-provider";

// Um atalho por área visível na navegação, agrupado pelo módulo (INT-01 AC4,
// DSS-06). O cartão inteiro é o link; ícone e descrição vêm do item de
// navegação definido pelo módulo e só aparecem quando ele os fornece.

function primeiroNome(nome: string): string {
  return nome.trim().split(/\s+/)[0] ?? "";
}

export function Atalhos() {
  const { context } = useSession();
  const nome = context ? primeiroNome(context.user.name) : "";

  return (
    <>
      <PageHeader title="Início" description={nome ? `Olá, ${nome}.` : undefined} />
      <Conteudo />
    </>
  );
}

function Conteudo() {
  const { context } = useSession();
  if (!context) return <LoadingState shape="list" />;

  const secoes = visibleItems(navItems, context.permissions).filter(isNavSection);

  if (secoes.length === 0) {
    return (
      <EmptyState
        title="Sua conta ainda não tem acesso a nenhum módulo."
        description="Se você precisa usar alguma área do sistema, fale com a administração."
      />
    );
  }

  return (
    <nav aria-label="Atalhos das áreas" className="flex flex-col gap-8">
      {secoes.map((secao) => (
        <Secao key={secao.label} secao={secao} />
      ))}
    </nav>
  );
}

function Secao({ secao }: { secao: NavSection }) {
  return (
    <section className="flex flex-col gap-4">
      <SectionTitle>{secao.label}</SectionTitle>
      <ul data-slot="atalhos-grade" className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {secao.items.map((item) => (
          <li key={item.href} className="flex">
            <Atalho item={item} />
          </li>
        ))}
      </ul>
    </section>
  );
}

function Atalho({ item }: { item: NavItem }) {
  const Icon = item.icon;
  return (
    <Card className="relative w-full gap-2 transition-colors duration-fast hover:border-border-strong hover:bg-accent">
      <div className="flex items-center gap-2">
        {Icon ? <Icon aria-hidden="true" className="size-5 shrink-0 text-muted-foreground" /> : null}
        {/* Link esticado: o cartão inteiro é clicável e o nome do link é só o nome da área. */}
        <Link
          href={item.href}
          className="text-base font-semibold outline-none after:absolute after:inset-0 after:rounded-lg focus-visible:after:ring-2 focus-visible:after:ring-ring focus-visible:after:ring-offset-2 focus-visible:after:ring-offset-ring-offset"
        >
          {item.label}
        </Link>
      </div>
      {item.description ? <p className="text-sm text-pretty text-muted-foreground">{item.description}</p> : null}
    </Card>
  );
}
