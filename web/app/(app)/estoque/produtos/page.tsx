import type { Metadata } from "next";

import { ProdutosPage } from "@/features/estoque/produtos/produtos-page";

export const metadata: Metadata = {
  title: "Produtos",
};

export default function EstoqueProdutosPage() {
  return (
    <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col px-4 py-8">
      <ProdutosPage />
    </main>
  );
}
