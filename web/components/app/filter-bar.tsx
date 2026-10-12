"use client";

import { SlidersHorizontalIcon } from "lucide-react";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetFooter,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useIsNarrow } from "@/lib/responsive";
import { cn } from "@/lib/utils";

// Barra de filtros (DSP-05 AC1). Não conhece a URL: o estado dos filtros é de
// quem usa, e vale na hora (não há rascunho a aplicar). A partir de 768 px os
// filtros ficam na barra; abaixo disso só a busca fica visível e os demais vão
// para uma gaveta inferior. A contagem fica numa região viva, para o efeito do
// filtro ser anunciado.
//
// Na barra, os campos se distribuem pela largura do próprio contêiner
// (container query), e não pela da tela: a mesma barra serve à página inteira,
// ao lado da navegação fixa ou numa região estreita. Contêiner estreito: um
// campo por linha; a partir de 28rem, duas colunas com a busca na linha toda;
// a partir de 48rem, colunas de pelo menos 11rem na mesma linha enquanto
// couberem, com a busca ocupando duas.

export type FilterBarProps = {
  /** Os filtros: `Field` com `NativeSelect` ou `Input` de data. */
  children: ReactNode;
  /** Campo de busca: fica sempre visível, também em tela estreita. */
  search?: ReactNode;
  /** Quantos filtros estão ativos; a busca preenchida conta. */
  activeCount: number;
  resultCount: number;
  resultNoun: readonly [singular: string, plural: string];
  onClear: () => void;
  className?: string;
};

const NUMBER = new Intl.NumberFormat("pt-BR");

const TITLE = "Filtros";
const CLEAR = "Limpar filtros";

export function FilterBar({
  children,
  search,
  activeCount,
  resultCount,
  resultNoun: [singular, plural],
  onClear,
  className,
}: FilterBarProps) {
  const narrow = useIsNarrow();
  const hasActive = activeCount > 0;

  return (
    <div data-slot="filter-bar" className={cn("@container flex w-full min-w-0 flex-col gap-2", className)}>
      {narrow ? (
        <div role="group" aria-label={TITLE} className="flex flex-col gap-3 rounded-md bg-muted p-3">
          {search}
          <Sheet>
            <SheetTrigger render={<Button variant="secondary" />}>
              <SlidersHorizontalIcon aria-hidden />
              {hasActive ? `${TITLE} (${activeCount})` : TITLE}
            </SheetTrigger>
            <SheetContent side="bottom">
              <SheetHeader>
                <SheetTitle>{TITLE}</SheetTitle>
              </SheetHeader>
              <div className="flex flex-col gap-4 px-4">{children}</div>
              <SheetFooter>
                {hasActive ? (
                  <Button variant="secondary" onClick={onClear}>
                    {CLEAR}
                  </Button>
                ) : null}
                <SheetClose render={<Button />}>Aplicar</SheetClose>
              </SheetFooter>
            </SheetContent>
          </Sheet>
        </div>
      ) : (
        <div
          role="group"
          aria-label={TITLE}
          className="grid grid-cols-1 items-end gap-3 rounded-md bg-muted p-3 *:min-w-0 @md:grid-cols-2 @3xl:grid-cols-[repeat(auto-fit,minmax(min(11rem,100%),1fr))]"
        >
          {search ? (
            <div data-slot="filter-bar-search" className="@md:col-span-2">
              {search}
            </div>
          ) : null}
          {children}
          {hasActive ? (
            <Button variant="ghost" className="justify-self-start" onClick={onClear}>
              {CLEAR}
            </Button>
          ) : null}
        </div>
      )}
      <p aria-live="polite" className="text-sm text-muted-foreground tabular-nums">
        {NUMBER.format(resultCount)} {resultCount === 1 ? singular : plural}
      </p>
    </div>
  );
}
