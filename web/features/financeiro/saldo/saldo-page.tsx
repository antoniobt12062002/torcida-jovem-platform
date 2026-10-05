"use client";

import { ErrorState, LoadingState } from "@/components/app/states";
import { formatBRL } from "@/lib/money";
import { cn } from "@/lib/utils";

import { useSaldo } from "./hooks";

// Saldo atual do financeiro (FWB-02): valor da API, formatado, e explicação
// de como ele é calculado. Nenhum cálculo é feito no front.

export const PERMISSION_SALDO_READ = "financeiro:saldo:read";

export function SaldoPage() {
  const query = useSaldo();

  return (
    <div className="flex flex-col gap-6">
      <h1 className="text-xl font-semibold">Saldo</h1>

      {query.isPending ? (
        <LoadingState label="Carregando saldo…" />
      ) : query.isError ? (
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      ) : (
        <section aria-label="Saldo atual" className="flex flex-col gap-2 rounded-lg border p-6">
          <p className="text-sm text-muted-foreground">Saldo atual</p>
          <p
            className={cn(
              "text-3xl font-semibold tabular-nums",
              query.data.saldo_cents < 0 && "text-destructive",
            )}
          >
            {formatBRL(query.data.saldo_cents)}
          </p>
        </section>
      )}

      <p className="text-sm text-muted-foreground">
        O saldo é a soma dos lançamentos recebidos menos a soma dos lançamentos pagos, pelo status atual de
        cada lançamento.
      </p>
    </div>
  );
}
