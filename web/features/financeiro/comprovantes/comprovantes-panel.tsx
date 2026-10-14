"use client";

import { DownloadIcon, PaperclipIcon } from "lucide-react";
import { type ChangeEvent, type RefObject, useRef, useState } from "react";

import { ApiErrorAlert } from "@/components/app/api-error";
import { type Column, DataList } from "@/components/app/data-list";
import { authorLabel, formatDateTime } from "@/components/app/format";
import { RequirePermission } from "@/components/app/require-permission";
import { RowActionButton } from "@/components/app/row-actions";
import { EmptyState, ErrorState, LoadingState } from "@/components/app/states";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { SectionTitle } from "@/components/ui/typography";
import { type ApiError, isApiError, networkError } from "@/lib/api/problem";
import { useSession } from "@/lib/session";

import { comprovantesCatalog, MESSAGE_ARQUIVO_ACIMA_DO_LIMITE } from "./errors";
import { type Comprovante, downloadComprovante, useComprovantes, useUploadComprovante } from "./hooks";
import { ACCEPTED_EXTENSIONS, formatBytes, formatLimit, MAX_UPLOAD_BYTES, withinUploadLimit } from "./limits";

// Painel de comprovantes no detalhe do lançamento (FWB-05): listar, anexar e
// baixar pela URL assinada, pedida só no clique em "Baixar". Com seis colunas
// a lista não cabe em 320 px como tabela (DSP-05 AC8): abaixo de 768 px vira
// lista empilhada, como as demais.

const TYPE_LABEL: Readonly<Record<string, string>> = {
  "application/pdf": "PDF",
  "image/jpeg": "JPEG",
  "image/png": "PNG",
  "image/webp": "WEBP",
};

const CREATE_PERMISSION = "financeiro:comprovante:create";

function newestFirst(items: readonly Comprovante[]): Comprovante[] {
  return [...items].sort((a, b) => Date.parse(b.uploaded_at) - Date.parse(a.uploaded_at));
}

export function ComprovantesPanel({ lancamentoId }: { lancamentoId: string }) {
  const query = useComprovantes(lancamentoId);
  const { context } = useSession();
  const [error, setError] = useState<ApiError | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);

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

  const columns: Column<Comprovante>[] = [
    { id: "nome", header: "Nome", priority: 1, inList: "title", wrap: "wrap", cell: (c) => c.original_filename },
    { id: "tipo", header: "Tipo", priority: 2, inList: "field", cell: (c) => TYPE_LABEL[c.content_type] ?? c.content_type },
    { id: "tamanho", header: "Tamanho", numeric: true, priority: 3, inList: false, cell: (c) => formatBytes(c.size_bytes) },
    { id: "versao", header: "Versão", numeric: true, priority: 3, inList: false, cell: (c) => c.version },
    {
      id: "autor",
      header: "Enviado por",
      priority: 2,
      inList: "field",
      cell: (c) => authorLabel(c.uploaded_by, context?.user.id),
    },
    {
      id: "data",
      header: "Enviado em",
      priority: 2,
      inList: "field",
      wrap: "nowrap",
      cell: (c) => formatDateTime(c.uploaded_at),
    },
  ];

  return (
    <section className="flex flex-col gap-4" aria-labelledby="comprovantes-titulo">
      <SectionTitle id="comprovantes-titulo">Comprovantes</SectionTitle>

      <RequirePermission permission={CREATE_PERMISSION} mode="action">
        <UploadComprovante
          lancamentoId={lancamentoId}
          inputRef={fileRef}
          onStart={() => setError(null)}
          onError={report}
        />
      </RequirePermission>

      {error ? <ApiErrorAlert error={error} catalog={comprovantesCatalog} /> : null}

      {query.isPending ? (
        <LoadingState shape="list" />
      ) : query.isError ? (
        <ErrorState error={query.error} catalog={comprovantesCatalog} onRetry={() => void query.refetch()} />
      ) : query.data.items.length === 0 ? (
        <EmptyState
          headingLevel={3}
          title="Nenhum comprovante anexado."
          description="Recibos e notas deste lançamento aparecem aqui."
          action={
            <RequirePermission permission={CREATE_PERMISSION} mode="action">
              <Button variant="secondary" onClick={() => fileRef.current?.click()}>
                <PaperclipIcon aria-hidden="true" />
                Escolher arquivo
              </Button>
            </RequirePermission>
          }
        />
      ) : (
        <DataList
          label="Comprovantes"
          columns={columns}
          rows={newestFirst(query.data.items)}
          rowKey={(c) => c.id}
          rowLabel={(c) => c.original_filename}
          primaryAction={(c) => (
            <RowActionButton
              label="Baixar"
              rowLabel={c.original_filename}
              icon={DownloadIcon}
              onClick={() => void download(c.id)}
            />
          )}
        />
      )}
    </section>
  );
}

function UploadComprovante({
  lancamentoId,
  inputRef,
  onStart,
  onError,
}: {
  lancamentoId: string;
  inputRef: RefObject<HTMLInputElement | null>;
  onStart: () => void;
  onError: (cause: unknown) => void;
}) {
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
      <Field invalid={tooLarge}>
        <FieldLabel>Anexar comprovante</FieldLabel>
        <Input
          ref={inputRef}
          type="file"
          name="comprovante"
          // O campo de arquivo nativo não centraliza o texto na vertical.
          className="py-1.5"
          accept={ACCEPTED_EXTENSIONS}
          disabled={upload.isPending}
          onChange={(event) => void onChange(event)}
        />
        <FieldDescription>PDF, JPG, JPEG, PNG ou WEBP, até {formatLimit(MAX_UPLOAD_BYTES)}.</FieldDescription>
      </Field>
      {upload.isPending ? (
        <p role="status" className="text-sm">
          Enviando…
        </p>
      ) : null}
      {tooLarge ? (
        <Alert variant="error">
          <AlertDescription>{MESSAGE_ARQUIVO_ACIMA_DO_LIMITE}</AlertDescription>
        </Alert>
      ) : null}
    </div>
  );
}
