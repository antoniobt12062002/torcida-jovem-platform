"use client";

import type { ReactNode } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { type ErrorCatalog, isApiError } from "@/lib/api/problem";
import { cn } from "@/lib/utils";

import { apiErrorMessage } from "./api-error";

// Estados padronizados das telas (FND-05 AC3).

export const MESSAGE_LOAD_FAILED = "Não foi possível carregar os dados.";

export function LoadingState({ label = "Carregando…", className }: { label?: string; className?: string }) {
  return (
    <div role="status" aria-live="polite" className={cn("flex flex-col gap-3 py-6", className)}>
      <Skeleton className="h-6 w-1/3" />
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-full" />
      <Skeleton className="h-4 w-2/3" />
      <span className="sr-only">{label}</span>
    </div>
  );
}

export type EmptyStateProps = {
  title?: string;
  description?: ReactNode;
  /** Ação principal, por exemplo o botão de criar o primeiro item. */
  action?: ReactNode;
  className?: string;
};

export function EmptyState({ title = "Nenhum item encontrado.", description, action, className }: EmptyStateProps) {
  return (
    <div
      className={cn(
        "flex flex-col items-center gap-2 rounded-lg border border-dashed px-4 py-10 text-center",
        className,
      )}
    >
      <p className="font-medium">{title}</p>
      {description ? <p className="text-sm text-muted-foreground">{description}</p> : null}
      {action ? <div className="mt-2">{action}</div> : null}
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
      <Alert variant="destructive">
        <AlertDescription>{message}</AlertDescription>
      </Alert>
      <Button variant="outline" onClick={onRetry}>
        Tentar de novo
      </Button>
    </div>
  );
}
