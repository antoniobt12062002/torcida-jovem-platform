import type { Metadata } from "next";

import { RequirePermission } from "@/components/app/require-permission";
import { SaldoPage } from "@/features/financeiro/saldo/saldo-page";

export const metadata: Metadata = {
  title: "Saldo",
};

export default function FinanceiroSaldoPage() {
  return (
    <main className="mx-auto flex w-full max-w-3xl flex-1 flex-col px-4 py-8">
      <RequirePermission permission="financeiro:saldo:read" mode="page">
        <SaldoPage />
      </RequirePermission>
    </main>
  );
}
