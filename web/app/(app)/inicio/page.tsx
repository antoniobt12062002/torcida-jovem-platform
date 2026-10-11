import type { Metadata } from "next";

import { Atalhos } from "./atalhos";

export const metadata: Metadata = {
  title: "Início",
};

// Página inicial autenticada (INT-01 AC4): atalhos para as áreas que a pessoa
// pode ver, gerados da mesma lista da navegação. Sem indicadores (fora da V1).
export default function InicioPage() {
  return (
    <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col gap-6 px-4 py-8">
      <h1 className="text-xl font-semibold">Início</h1>
      <Atalhos />
    </main>
  );
}
