"use client";

import { MoneyValue } from "@/components/app/money-value";
import { PageHeader } from "@/components/app/page-header";
import { ErrorState, LoadingState } from "@/components/app/states";
import { Card } from "@/components/ui/card";

import { useSaldo } from "./hooks";

// Saldo atual do financeiro (FWB-02): valor da API e explicação de como ele é
// calculado. Nenhum cálculo é feito no front. O valor é o saldo isolado de
// destaque (DD-4, WDS-D-031): `MoneyValue` `balance` em `display`, que no
// negativo mostra o sinal de menos, o ícone e o texto "saldo negativo".

export const PERMISSION_SALDO_READ = "financeiro:saldo:read";

export const EXPLICACAO =
  "O saldo é a soma dos lançamentos recebidos menos a soma dos lançamentos pagos, pelo status atual de cada lançamento.";

export const AVISO_NEGATIVO = "Os lançamentos pagos superam os recebidos.";

export function SaldoPage() {
  const query = useSaldo();

  return (
    <>
      <PageHeader title="Saldo" />

      {query.isPending ? (
        <LoadingState label="Carregando saldo…" shape="value" />
      ) : query.isError ? (
        <ErrorState error={query.error} onRetry={() => void query.refetch()} />
      ) : (
        <Card aria-label="Saldo atual" role="region" className="gap-2">
          <p className="text-sm text-muted-foreground">Saldo atual</p>
          <MoneyValue cents={query.data.saldo_cents} variant="balance" size="display" className="self-start" />
          {query.data.saldo_cents < 0 ? <p className="text-sm text-destructive">{AVISO_NEGATIVO}</p> : null}
        </Card>
      )}

      <p className="text-sm text-muted-foreground">{EXPLICACAO}</p>
    </>
  );
}
