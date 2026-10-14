import type { Metadata } from "next";

import { PageContainer } from "@/components/app/page-container";
import { NoAccess } from "@/components/app/require-permission";

export const metadata: Metadata = {
  title: "Sem acesso",
};

// FND-04 AC4: área sem a permissão de leitura. Dentro do shell, com o estado
// padrão "Sem acesso" (DSP-02 AC5), que traz o único `h1` da página.
export default function SemAcessoPage() {
  return (
    <PageContainer width="narrow">
      <NoAccess />
    </PageContainer>
  );
}
