"use client";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { type ApiError, type ErrorCatalog, isUnavailable, messageFor } from "@/lib/api/problem";

// Apresentação dos erros da API (FND-04 AC5, FND-06 AC2-5).

export const MESSAGE_FORBIDDEN = "Você não tem permissão para esta ação.";

/**
 * Mensagem em português para a pessoa: o catálogo da tela primeiro, depois
 * `403 forbidden` e as mensagens genéricas de `messageFor`.
 */
export function apiErrorMessage(error: ApiError, catalog: ErrorCatalog = {}): string {
  if (!Object.hasOwn(catalog, error.code) && error.status === 403 && error.code === "forbidden") {
    return MESSAGE_FORBIDDEN;
  }
  return messageFor(error, catalog);
}

export type ApiErrorAlertProps = {
  error: ApiError;
  catalog?: ErrorCatalog;
  /** Mostrado só quando a API está indisponível (503 ou rede). */
  onRetry?: () => void;
  className?: string;
};

export function ApiErrorAlert({ error, catalog, onRetry, className }: ApiErrorAlertProps) {
  return (
    <Alert variant="destructive" className={className}>
      <AlertDescription className="flex flex-col items-start gap-2">
        <span>{apiErrorMessage(error, catalog)}</span>
        {onRetry && isUnavailable(error) ? (
          <Button variant="outline" size="sm" onClick={onRetry}>
            Tentar de novo
          </Button>
        ) : null}
      </AlertDescription>
    </Alert>
  );
}
