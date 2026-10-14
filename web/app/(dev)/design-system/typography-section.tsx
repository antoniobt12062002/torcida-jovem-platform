import type { ReactNode } from "react";

import { MoneyValue } from "@/components/app/money-value";
import { DisplayTitle, Eyebrow } from "@/components/ui/typography";

import { SpecimenSection } from "./section";

// Tipografia (DSF-03): cada papel da escala com texto real em português, com
// acentos e caixa alta. O título de página e o de seção são os da própria
// página (um só `h1`; as seções em `h2`), por isso não se repetem aqui.

type Role = { name: string; spec: string; sample: ReactNode };

const ROLES: readonly Role[] = [
  {
    name: "Destaque",
    spec: "DisplayTitle · Barlow Condensed 700 · 36/40 · caixa alta",
    sample: <DisplayTitle>Ação, união e paixão pelo futsal</DisplayTitle>,
  },
  {
    name: "Título de página",
    spec: "PageTitle (h1) · Barlow Condensed 700 · 30/36; 24/28 abaixo de 768 px · caixa alta, com o filete de marca",
    sample: <p className="text-sm">É o título no topo desta página, &quot;Espécime do design system&quot;.</p>,
  },
  {
    name: "Título de seção",
    spec: "SectionTitle (h2) · Barlow Condensed 600 · 20/24 · frase comum, com a régua forte",
    sample: <p className="text-sm">É o título de cada seção desta página, como &quot;Tipografia&quot; acima.</p>,
  },
  {
    name: "Rótulo de grupo e sobretítulo",
    spec: "Eyebrow · Barlow Condensed 600 · 14/16 · caixa alta",
    sample: <Eyebrow>Gestão financeira e prestação de contas</Eyebrow>,
  },
  {
    name: "Título de bloco",
    spec: "h3 e título de diálogo · Geist 600 · 16/24",
    sample: <p className="text-base leading-6 font-semibold">Lançamentos de março e situação do caixa</p>,
  },
  {
    name: "Corpo",
    spec: "Geist 400 · 14/20",
    sample: (
      <p className="text-sm">
        A prestação de contas é publicada após a reunião do conselho, com as receitas, as despesas e o saldo de cada mês.
      </p>
    ),
  },
  {
    name: "Corpo maior",
    spec: "Geist 400 · 16/24",
    sample: (
      <p className="text-base">Associação, eventos e loja ficam no mesmo lugar, com histórico e auditoria de cada mudança.</p>
    ),
  },
  {
    name: "Rótulo",
    spec: "Rótulo de campo e botão · Geist 500 · 14/20",
    sample: <p className="text-sm font-medium">Descrição do lançamento</p>,
  },
  {
    name: "Legenda",
    spec: "Ajuda e metadado · Geist 400 · 12/16",
    sample: <p className="text-xs text-muted-foreground">Última alteração em 14/03/2026, às 19h40, por Tesouraria.</p>,
  },
  {
    name: "Dado",
    spec: "Número e dinheiro · Geist 400 ou 500 · 14/20 · algarismos tabulares",
    sample: (
      <p className="flex flex-wrap gap-4 text-sm tabular-nums">
        <span>1.234 associados</span>
        <MoneyValue cents={123456} />
      </p>
    ),
  },
  {
    name: "Valor de destaque",
    spec: "MoneyValue size=display · fonte do token --font-money-display (Barlow Condensed 700, DD-4) · 36/40",
    sample: <MoneyValue cents={1847390} variant="balance" size="display" />,
  },
  {
    name: "Identificador",
    spec: "Código curto · Geist Mono 400 · 14/20",
    sample: <p className="font-mono text-sm">fb6e3d8f · CMF-2026-0314</p>,
  },
];

export function TypographySection() {
  return (
    <SpecimenSection
      id="tipografia"
      title="Tipografia"
      description="Geist no corpo e nos dados; Barlow Condensed só nos papéis de identidade. Escala padrão do Tailwind (12, 14, 16, 18, 20, 24, 30 e 36 px), sem valor arbitrário."
    >
      <dl className="flex flex-col divide-y border-y">
        {ROLES.map((role) => (
          <div key={role.name} className="flex flex-col gap-2 py-4 md:flex-row md:gap-6">
            <dt className="flex flex-col gap-1 md:w-1/3 md:shrink-0">
              <span className="text-sm font-medium">{role.name}</span>
              <span className="text-xs text-muted-foreground">{role.spec}</span>
            </dt>
            <dd className="min-w-0 flex-1">{role.sample}</dd>
          </div>
        ))}
      </dl>
    </SpecimenSection>
  );
}
