import type { Metadata } from "next";
import { Suspense } from "react";

import { PageContainer } from "@/components/app/page-container";
import { LoadingState } from "@/components/app/states";
import { ProdutosPage } from "@/features/estoque/produtos/produtos-page";

export const metadata: Metadata = {
  title: "Produtos",
};

// A busca fica na URL (`useSearchParams`), por isso a lista fica dentro de `Suspense`.
export default function EstoqueProdutosPage() {
  return (
    <PageContainer>
      <Suspense fallback={<LoadingState shape="table" />}>
        <ProdutosPage />
      </Suspense>
    </PageContainer>
  );
}
