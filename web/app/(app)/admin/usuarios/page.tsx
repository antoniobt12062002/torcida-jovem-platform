import type { Metadata } from "next";
import { Suspense } from "react";

import { PageContainer } from "@/components/app/page-container";
import { RequirePermission } from "@/components/app/require-permission";
import { LoadingState } from "@/components/app/states";
import { UsersPage } from "@/features/identity/usuarios/users-page";

export const metadata: Metadata = {
  title: "Usuários",
};

// Os filtros da lista ficam na URL: `useSearchParams` exige o `<Suspense>`.
export default function UsuariosPage() {
  return (
    <PageContainer width="wide">
      <RequirePermission permission="identity:user:read" mode="page">
        <Suspense fallback={<LoadingState shape="table" />}>
          <UsersPage />
        </Suspense>
      </RequirePermission>
    </PageContainer>
  );
}
