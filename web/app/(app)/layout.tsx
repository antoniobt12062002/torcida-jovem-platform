"use client";

import { usePathname, useRouter } from "next/navigation";
import { type ReactNode, useEffect } from "react";

import { AppShell } from "@/components/app/app-shell";
import { navItems } from "@/components/app/nav-items";
import { PageContainer } from "@/components/app/page-container";
import { ErrorState, LoadingState } from "@/components/app/states";
import { routes } from "@/lib/session/routes";
import { useSession } from "@/lib/session/session-provider";

// Portão das rotas autenticadas (FND-03 AC4, FND-04 AC1-2, WEB-D-009). É só
// navegação: a API continua sendo a autoridade.

export default function AppLayout({ children }: { children: ReactNode }) {
  const { status, context, error, refetch, redirectToLogin } = useSession();
  const pathname = usePathname();
  const router = useRouter();

  const mustChangePassword =
    status === "authenticated" && !!context?.must_change_password && pathname !== routes.contaSenha;

  useEffect(() => {
    if (status === "anonymous") redirectToLogin();
  }, [status, redirectToLogin]);

  useEffect(() => {
    if (mustChangePassword) router.replace(routes.contaSenha);
  }, [mustChangePassword, router]);

  if (status === "authenticated" && !mustChangePassword) return <AppShell items={navItems}>{children}</AppShell>;

  // Sem o shell: o contêiner da página traz a única região principal (DSP-02, DSS-01 AC6).
  if (status === "loading" && error) {
    return (
      <PageContainer width="narrow">
        <ErrorState error={error} onRetry={refetch} />
      </PageContainer>
    );
  }

  return (
    <PageContainer width="narrow">
      <LoadingState />
    </PageContainer>
  );
}
