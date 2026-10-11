import type { Metadata } from "next";

import { ProdutoDetail } from "@/features/estoque/produto-detail";

export const metadata: Metadata = {
  title: "Produto",
};

export default async function EstoqueProdutoPage({ params }: PageProps<"/estoque/produtos/[id]">) {
  const { id } = await params;
  return (
    <main className="mx-auto flex w-full max-w-5xl flex-1 flex-col px-4 py-8">
      <ProdutoDetail id={id} />
    </main>
  );
}
