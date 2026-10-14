import type { Metadata } from "next";

import { PageContainer } from "@/components/app/page-container";
import { ProdutoDetail } from "@/features/estoque/produto-detail";

export const metadata: Metadata = {
  title: "Produto",
};

export default async function EstoqueProdutoPage({ params }: PageProps<"/estoque/produtos/[id]">) {
  const { id } = await params;
  return (
    <PageContainer>
      <ProdutoDetail id={id} />
    </PageContainer>
  );
}
