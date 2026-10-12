import { cn } from "@/lib/utils";

import { SpecimenBlock, SpecimenCaption, SpecimenSection } from "./section";

// Espaço, raio, borda e sombra (DSF-04): amostras com o nome do token. As
// medidas dos tokens de componente vêm de `globals.css`.

// Passos permitidos da base de 4 px, com a classe escrita por inteiro.
const SPACING: readonly { step: number; className: string }[] = [
  { step: 1, className: "w-1" },
  { step: 2, className: "w-2" },
  { step: 3, className: "w-3" },
  { step: 4, className: "w-4" },
  { step: 5, className: "w-5" },
  { step: 6, className: "w-6" },
  { step: 8, className: "w-8" },
  { step: 10, className: "w-10" },
  { step: 12, className: "w-12" },
  { step: 16, className: "w-16" },
];

const RADII: readonly { token: string; className: string; use: string }[] = [
  { token: "rounded-sm (--radius-sm)", className: "rounded-sm", use: "Badge retangular, caixa de seleção, foco de link" },
  { token: "rounded-md (--radius-md)", className: "rounded-md", use: "Botões, campos, itens de menu" },
  { token: "rounded-lg (--radius-lg)", className: "rounded-lg", use: "Cartões, diálogos, toasts, menus" },
  { token: "rounded-full", className: "rounded-full", use: "Só indicador de status" },
];

const BORDERS: readonly { name: string; className: string; use: string }[] = [
  { name: "Decorativa · 1 px border", className: "border border-border", use: "Cartão, divisória, linha de tabela" },
  { name: "De controle · 1 px input", className: "border border-input", use: "Campos" },
  { name: "Forte · 2 px border-strong", className: "border-b-2 border-border-strong", use: "Régua de tabela e de título de seção" },
  { name: "De marca · 2 px brand", className: "border-b-2 border-brand", use: "Filete do título de página" },
  { name: "De erro · 1 px destructive", className: "border border-destructive", use: "Campo inválido" },
];

const SHADOWS: readonly { name: string; className: string; use: string }[] = [
  { name: "Nível 0 · nenhuma", className: "border", use: "Tudo em repouso: cartões, tabelas, campos, shell" },
  { name: "Nível 1 · shadow-elevation-1", className: "shadow-elevation-1", use: "Menu, popover, toast" },
  { name: "Nível 2 · shadow-elevation-2", className: "shadow-elevation-2", use: "Diálogo, gaveta" },
];

// Tokens de componente com a amostra de altura (classe por inteiro).
const COMPONENT_TOKENS: readonly { name: string; className?: string }[] = [
  { name: "control-height-sm", className: "h-control-sm" },
  { name: "control-height-md", className: "h-control-md" },
  { name: "control-height-lg", className: "h-control-lg" },
  { name: "touch-target", className: "h-touch" },
  { name: "table-head-height", className: "h-table-head" },
  { name: "table-row-height", className: "h-table-row" },
  { name: "shell-header-height", className: "h-shell-header" },
  { name: "control-font-size" },
  { name: "shell-nav-width" },
  { name: "content-narrow" },
  { name: "content-default" },
  { name: "content-wide" },
  { name: "page-gutter" },
];

const SAMPLE_GRID = "grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3";

export function SpaceSection({ declarations }: { declarations: Record<string, string> }) {
  return (
    <SpecimenSection
      id="espaco-raio-borda-sombra"
      title="Espaço, raio, borda e sombra"
      description="Profundidade no conteúdo vem de borda e de fundo muted; sombra só em sobreposições. Sem gradiente e sem brilho."
    >
      <SpecimenBlock title="Espaçamento (base de 4 px)">
        <ul className="flex flex-col gap-2">
          {SPACING.map(({ step, className }) => (
            <li key={step} className="flex items-center gap-3">
              <span className="w-24 shrink-0 font-mono text-sm">spacing-{step}</span>
              <span aria-hidden="true" className={cn("h-4 shrink-0 bg-foreground", className)} />
              <span className="text-xs text-muted-foreground">{step * 4} px</span>
            </li>
          ))}
        </ul>
      </SpecimenBlock>

      <SpecimenBlock title="Raio">
        <ul className={SAMPLE_GRID}>
          {RADII.map(({ token, className, use }) => (
            <li key={token} className="flex items-center gap-3">
              <span aria-hidden="true" className={cn("size-12 shrink-0 border-2 border-border-strong bg-muted", className)} />
              <span className="flex flex-col">
                <span className="font-mono text-sm">{token}</span>
                <SpecimenCaption>{use}</SpecimenCaption>
              </span>
            </li>
          ))}
        </ul>
      </SpecimenBlock>

      <SpecimenBlock title="Borda">
        <ul className={SAMPLE_GRID}>
          {BORDERS.map(({ name, className, use }) => (
            <li key={name} className="flex flex-col gap-2">
              <span aria-hidden="true" className={cn("h-12 w-full bg-background", className)} />
              <span className="text-sm font-medium">{name}</span>
              <SpecimenCaption>{use}</SpecimenCaption>
            </li>
          ))}
        </ul>
      </SpecimenBlock>

      <SpecimenBlock title="Sombra">
        <ul className={SAMPLE_GRID}>
          {SHADOWS.map(({ name, className, use }) => (
            <li key={name} className="flex flex-col gap-2 p-2">
              <span aria-hidden="true" className={cn("h-16 w-full rounded-lg bg-popover", className)} />
              <span className="text-sm font-medium">{name}</span>
              <SpecimenCaption>{use}</SpecimenCaption>
            </li>
          ))}
        </ul>
      </SpecimenBlock>

      <SpecimenBlock title="Tokens de componente">
        <p className="text-sm text-muted-foreground">
          Valores de app/globals.css fora de ponteiro de toque. Sob pointer: coarse os controles passam a 44 px e o texto de campo a 16 px.
        </p>
        <ul className="flex flex-col divide-y border-y">
          {COMPONENT_TOKENS.map(({ name, className }) => (
            <li key={name} className="flex min-h-control-lg flex-wrap items-center gap-3 py-2">
              <span className="w-48 shrink-0 font-mono text-sm">--{name}</span>
              <span className="text-sm tabular-nums">{declarations[name] ?? "sem valor"}</span>
              {className ? <span aria-hidden="true" className={cn("w-4 bg-foreground", className)} /> : null}
            </li>
          ))}
        </ul>
      </SpecimenBlock>
    </SpecimenSection>
  );
}
