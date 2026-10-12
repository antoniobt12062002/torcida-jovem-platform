"use client";

import { InboxIcon, type LucideIcon, SearchXIcon } from "lucide-react";
import type { ReactNode } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { type ErrorCatalog, isApiError } from "@/lib/api/problem";
import { cn } from "@/lib/utils";

import { apiErrorMessage } from "./api-error";

// Estados padronizados das telas (FND-05 AC3, DSP-02). Ocupam o lugar do
// conteúdo: o cabeçalho de página e a barra de filtros continuam onde estão.

export const MESSAGE_LOAD_FAILED = "Não foi possível carregar os dados.";

/** Formato aproximado do conteúdo que está carregando. */
export type LoadingShape = "table" | "list" | "detail" | "value";

function LoadingSkeleton({ shape }: { shape?: LoadingShape }) {
  switch (shape) {
    case "table":
      return (
        <>
          <Skeleton className="h-table-head w-full" />
          <Skeleton shape="row" />
          <Skeleton shape="row" />
          <Skeleton shape="row" />
          <Skeleton shape="row" />
        </>
      );
    case "list":
      return (
        <>
          {[0, 1, 2].map((item) => (
            <div key={item} className="flex flex-col gap-2">
              <Skeleton className="h-5 w-1/2" />
              <Skeleton shape="text" className="w-3/4" />
            </div>
          ))}
        </>
      );
    case "detail":
      return (
        <>
          <Skeleton className="h-8 w-1/2" />
          <Skeleton shape="text" className="w-1/3" />
          <Skeleton shape="text" />
          <Skeleton shape="text" />
          <Skeleton shape="text" className="w-2/3" />
        </>
      );
    case "value":
      return <Skeleton shape="block" />;
    default:
      return (
        <>
          <Skeleton className="h-6 w-1/3" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-full" />
          <Skeleton className="h-4 w-2/3" />
        </>
      );
  }
}

export type LoadingStateProps = {
  /** Texto anunciado a leitores de tela. Padrão: "Carregando…". */
  label?: string;
  /** Formato do esqueleto. Sem `shape`, um bloco genérico de texto. */
  shape?: LoadingShape;
  className?: string;
};

export function LoadingState({ label = "Carregando…", shape, className }: LoadingStateProps) {
  return (
    <div role="status" aria-live="polite" className={cn("flex flex-col gap-3 py-6", className)}>
      <LoadingSkeleton shape={shape} />
      <span className="sr-only">{label}</span>
    </div>
  );
}

export type EmptyKind = "empty" | "no-results";

const EMPTY: Record<EmptyKind, { icon: LucideIcon; title: string }> = {
  empty: { icon: InboxIcon, title: "Nenhum item encontrado." },
  "no-results": { icon: SearchXIcon, title: "Nenhum resultado para os filtros" },
};

export type EmptyStateProps = {
  /** `empty`: não há registro. `no-results`: os filtros ou a busca não acharam nada. Padrão: `empty`. */
  kind?: EmptyKind;
  /** Padrão: o título do tipo. */
  title?: string;
  /** Frase de orientação: o que fazer em seguida. */
  description?: ReactNode;
  /** Ação principal, por exemplo o botão de criar o primeiro item. */
  action?: ReactNode;
  /** Com `kind="no-results"` e sem `action`, mostra a ação "Limpar filtros". */
  onClear?: () => void;
  /** Ícone no lugar do padrão do tipo. */
  icon?: LucideIcon;
  /** Nível do título: 2 no lugar da lista da página; 3 dentro de uma seção. Padrão: 2. */
  headingLevel?: 2 | 3;
  className?: string;
};

export function EmptyState({
  kind = "empty",
  title,
  description,
  action,
  onClear,
  icon,
  headingLevel = 2,
  className,
}: EmptyStateProps) {
  const Icon = icon ?? EMPTY[kind].icon;
  const Heading = headingLevel === 2 ? "h2" : "h3";
  const shownAction =
    action ??
    (kind === "no-results" && onClear ? (
      <Button variant="secondary" onClick={onClear}>
        Limpar filtros
      </Button>
    ) : null);
  return (
    <div
      data-slot="empty-state"
      data-kind={kind}
      className={cn("flex flex-col items-center gap-2 px-4 py-8 text-center md:py-12", className)}
    >
      <Icon aria-hidden="true" className="size-6 text-muted-foreground" />
      <Heading className="text-base font-semibold text-balance">{title ?? EMPTY[kind].title}</Heading>
      {description ? <p className="text-sm text-pretty text-muted-foreground">{description}</p> : null}
      {shownAction ? <div className="mt-2">{shownAction}</div> : null}
    </div>
  );
}

export type ErrorStateProps = {
  /** Erro da consulta. Um ApiError usa o catálogo; outro erro vira mensagem genérica. */
  error: unknown;
  catalog?: ErrorCatalog;
  onRetry: () => void;
  className?: string;
};

export function ErrorState({ error, catalog, onRetry, className }: ErrorStateProps) {
  const message = isApiError(error) ? apiErrorMessage(error, catalog) : MESSAGE_LOAD_FAILED;
  return (
    <div className={cn("flex flex-col items-start gap-3 py-6", className)}>
      <Alert variant="error">
        <AlertDescription>{message}</AlertDescription>
      </Alert>
      <Button variant="secondary" onClick={onRetry}>
        Tentar de novo
      </Button>
    </div>
  );
}
