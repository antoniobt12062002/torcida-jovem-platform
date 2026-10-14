"use client";

import { CircleAlertIcon, CircleCheckIcon } from "lucide-react";

import { type Column, DataList } from "@/components/app/data-list";
import { Badge } from "@/components/ui/badge";
import { contrastRatio } from "@/lib/design/contrast";
import { CONTRAST_MINIMUM, CONTRAST_PAIRS, type ContrastPair } from "@/lib/design/contrast-pairs";
import { cn } from "@/lib/utils";

import { BG, COLOR_GROUP_ORDER, COLOR_GROUPS, type ColorGroup, groupOf, TEXT } from "./color-classes";
import { SpecimenBlock, SpecimenSection } from "./section";
import type { Palette } from "./tokens";

// Cores (DSF-06 AC1, DSF-02): primitivos, semânticos por grupo e cada par de
// `contrast-pairs.ts` com a razão calculada por `contrast.ts` a partir dos
// valores lidos de `globals.css`. Nenhum número escrito à mão.

const RATIO = new Intl.NumberFormat("pt-BR", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
const MINIMUM = new Intl.NumberFormat("pt-BR", { maximumFractionDigits: 1 });

const GROUP_LABEL: Record<ColorGroup, string> = {
  brand: "brand: marca",
  surface: "surface: superfícies e bordas",
  content: "content: texto",
  action: "action: ações",
  state: "state: estados",
  focus: "focus: foco",
};

type PairRow = ContrastPair & { ratio: number; passes: boolean };

function PairSample({ pair }: { pair: PairRow }) {
  if (pair.kind === "text") {
    return (
      <span
        className={cn(
          "inline-flex h-control-md items-center rounded-md border px-3 text-sm font-medium",
          BG[pair.background],
          TEXT[pair.foreground],
        )}
      >
        Aa Ação
      </span>
    );
  }
  return (
    <span className={cn("inline-flex h-control-md w-16 items-center justify-center rounded-md border", BG[pair.background])}>
      <span className={cn("h-1 w-8", BG[pair.foreground])} />
    </span>
  );
}

const PAIR_COLUMNS: readonly Column<PairRow>[] = [
  { id: "amostra", header: "Amostra", priority: 1, inList: "meta", cell: (pair) => <PairSample pair={pair} /> },
  {
    id: "par",
    header: "Par",
    priority: 1,
    inList: "title",
    cell: (pair) => (
      <span className="font-mono text-sm">
        {pair.foreground} sobre {pair.background}
      </span>
    ),
  },
  { id: "uso", header: "Uso", priority: 3, inList: "field", cell: (pair) => pair.use },
  {
    id: "tipo",
    header: "Tipo",
    priority: 2,
    inList: "field",
    cell: (pair) => `${pair.kind === "text" ? "Texto" : "Gráfico"}, mínimo ${MINIMUM.format(CONTRAST_MINIMUM[pair.kind])}:1`,
  },
  {
    id: "razao",
    header: "Razão",
    numeric: true,
    priority: 1,
    cell: (pair) => (
      <span className="inline-flex items-center gap-2">
        {RATIO.format(pair.ratio)}:1
        <Badge variant={pair.passes ? "success" : "error"}>
          {pair.passes ? <CircleCheckIcon aria-hidden="true" /> : <CircleAlertIcon aria-hidden="true" />}
          {pair.passes ? "Atinge" : "Abaixo"}
        </Badge>
      </span>
    ),
  },
];

function Swatch({ token, palette }: { token: string; palette: Palette }) {
  const entry = palette.semantic[token];
  return (
    <li className="flex items-center gap-3">
      <span aria-hidden="true" className={cn("size-10 shrink-0 rounded-md border", BG[token])} />
      <span className="flex min-w-0 flex-col">
        <span className="font-mono text-sm wrap-break-word">{token}</span>
        <span className="text-xs wrap-break-word text-muted-foreground">
          {entry ? `${entry.reference} · ${entry.value}` : "sem valor"}
        </span>
      </span>
    </li>
  );
}

const SWATCH_GRID = "grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4";

export function ColorsSection({ palette }: { palette: Palette }) {
  const rows: PairRow[] = CONTRAST_PAIRS.map((pair) => {
    const ratio = contrastRatio(palette.semantic[pair.foreground].value, palette.semantic[pair.background].value);
    return { ...pair, ratio, passes: ratio >= CONTRAST_MINIMUM[pair.kind] };
  });

  return (
    <SpecimenSection
      id="cores"
      title="Cores"
      description="Valores lidos de app/globals.css. Cada par de contraste vem de lib/design/contrast-pairs.ts, com a razão calculada por lib/design/contrast.ts (WCAG 2.2: 4,5:1 em texto, 3:1 em gráfico)."
    >
      <SpecimenBlock title="Primitivos">
        <p className="text-sm text-muted-foreground">
          A paleta oficial e as escalas derivadas. Não são expostos ao Tailwind: só a camada semântica os lê.
        </p>
        <ul className={SWATCH_GRID}>
          {palette.primitives.map(({ name, value }) => (
            <li key={name} className="flex items-center gap-3">
              {/* Primitivo não tem utilitário: a amostra lê a própria variável. */}
              <span
                aria-hidden="true"
                className="size-10 shrink-0 rounded-md border"
                style={{ backgroundColor: `var(--${name})` }}
              />
              <span className="flex min-w-0 flex-col">
                <span className="font-mono text-sm">--{name}</span>
                <span className="text-xs text-muted-foreground">{value}</span>
              </span>
            </li>
          ))}
        </ul>
      </SpecimenBlock>

      {COLOR_GROUP_ORDER.map((group) => {
        const pairs = rows.filter((pair) => groupOf(pair.foreground) === group);
        return (
          <SpecimenBlock key={group} title={`Semânticos ${GROUP_LABEL[group]}`}>
            <ul className={SWATCH_GRID}>
              {COLOR_GROUPS[group].map((token) => (
                <Swatch key={token} token={token} palette={palette} />
              ))}
            </ul>
            {pairs.length > 0 ? (
              <DataList
                label={`Pares de contraste do grupo ${group}`}
                columns={PAIR_COLUMNS}
                rows={pairs}
                rowKey={(pair) => `${pair.foreground}/${pair.background}/${pair.kind}`}
                rowLabel={(pair) => `${pair.foreground} sobre ${pair.background}`}
              />
            ) : null}
          </SpecimenBlock>
        );
      })}
    </SpecimenSection>
  );
}
