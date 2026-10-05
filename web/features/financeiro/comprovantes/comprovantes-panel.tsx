"use client";

import { type ChangeEvent, useId, useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { authorLabel, formatDateTime } from "@/components/app/format";
import { RequirePermission } from "@/components/app/require-permission";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "@/components/ui/toast";
import { type ApiError, isApiError, networkError } from "@/lib/api/problem";
import { useSession } from "@/lib/session";

import { comprovantesCatalog, MESSAGE_ARQUIVO_ACIMA_DO_LIMITE } from "./errors";
import { type Comprovante, downloadComprovante, useComprovantes, useUploadComprovante } from "./hooks";
import { ACCEPTED_EXTENSIONS, formatBytes, MAX_UPLOAD_BYTES, withinUploadLimit } from "./limits";

// Painel de comprovantes no detalhe do lançamento (FWB-05): listar, anexar e
// baixar pela URL assinada.

const TYPE_LABEL: Readonly<Record<string, string>> = {
  "application/pdf": "PDF",
  "image/jpeg": "JPEG",
  "image/png": "PNG",
  "image/webp": "WEBP",
};

function newestFirst(items: readonly Comprovante[]): Comprovante[] {
  return [...items].sort((a, b) => Date.parse(b.uploaded_at) - Date.parse(a.uploaded_at));
}

export function ComprovantesPanel({ lancamentoId }: { lancamentoId: string }) {
  const query = useComprovantes(lancamentoId);
  const { context } = useSession();
  const [error, setError] = useState<ApiError | null>(null);

  const report = (cause: unknown) => setError(isApiError(cause) ? cause : networkError(cause));

  async function download(documentId: string) {
    setError(null);
    try {
      const url = await downloadComprovante(documentId);
      window.open(url, "_blank", "noopener,noreferrer");
    } catch (cause) {
      report(cause);
    }
  }

  return (
    <section className="flex flex-col gap-4" aria-labelledby="comprovantes-titulo">
      <h2 id="comprovantes-titulo" className="text-lg font-semibold">
        Comprovantes
      </h2>

      <RequirePermission permission="financeiro:comprovante:create" mode="action">
        <UploadComprovante lancamentoId={lancamentoId} onStart={() => setError(null)} onError={report} />
      </RequirePermission>

      {error ? <ApiErrorAlert error={error} catalog={comprovantesCatalog} /> : null}

      {query.isPending ? (
        <LoadingState />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={comprovantesCatalog} onRetry={() => void query.refetch()} />
      ) : query.data.items.length === 0 ? (
        <EmptyState title="Nenhum comprovante anexado." />
      ) : (
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Nome</TableHead>
              <TableHead>Tipo</TableHead>
              <TableHead className="text-right">Tamanho</TableHead>
              <TableHead className="text-right">Versão</TableHead>
              <TableHead>Enviado por</TableHead>
              <TableHead>Enviado em</TableHead>
              <TableHead>
                <span className="sr-only">Baixar</span>
              </TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {newestFirst(query.data.items).map((c) => (
              <TableRow key={c.id} data-testid="comprovante-row">
                <TableCell>{c.original_filename}</TableCell>
                <TableCell>{TYPE_LABEL[c.content_type] ?? c.content_type}</TableCell>
                <TableCell className="text-right">{formatBytes(c.size_bytes)}</TableCell>
                <TableCell className="text-right">{c.version}</TableCell>
                <TableCell>{authorLabel(c.uploaded_by, context?.user.id)}</TableCell>
                <TableCell>{formatDateTime(c.uploaded_at)}</TableCell>
                <TableCell>
                  <Button
                    variant="outline"
                    size="sm"
                    aria-label={`Baixar ${c.original_filename}`}
                    onClick={() => void download(c.id)}
                  >
                    Baixar
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      )}
    </section>
  );
}

function UploadComprovante({
  lancamentoId,
  onStart,
  onError,
}: {
  lancamentoId: string;
  onStart: () => void;
  onError: (cause: unknown) => void;
}) {
  const id = useId();
  const upload = useUploadComprovante(lancamentoId);
  const [tooLarge, setTooLarge] = useState(false);

  async function onChange(event: ChangeEvent<HTMLInputElement>) {
    const input = event.currentTarget;
    const file = input.files?.[0];
    if (!file) return;
    onStart();
    setTooLarge(false);
    if (!withinUploadLimit(file.size)) {
      setTooLarge(true);
      input.value = "";
      return;
    }
    try {
      await upload.mutateAsync(file);
      toast.success("Comprovante anexado.");
    } catch (cause) {
      onError(cause);
    } finally {
      input.value = "";
    }
  }

  return (
    <div className="flex flex-col gap-2">
      <label htmlFor={id} className="text-sm font-medium">
        Anexar comprovante
      </label>
      <Input
        id={id}
        type="file"
        accept={ACCEPTED_EXTENSIONS}
        disabled={upload.isPending}
        aria-describedby={`${id}-dica`}
        onChange={(event) => void onChange(event)}
      />
      <p id={`${id}-dica`} className="text-sm text-muted-foreground">
        PDF, JPG, JPEG, PNG ou WEBP, até {formatBytes(MAX_UPLOAD_BYTES)}.
      </p>
      {upload.isPending ? (
        <p role="status" className="text-sm">
          Enviando…
        </p>
      ) : null}
      {tooLarge ? (
        <Alert variant="destructive">
          <AlertDescription>{MESSAGE_ARQUIVO_ACIMA_DO_LIMITE}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  );
}
