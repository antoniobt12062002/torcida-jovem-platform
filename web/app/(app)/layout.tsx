"use client";

import { usePathname, useRouter } from "next/navigation";
import { type ReactNode, useEffect } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { messageFor } from "@/lib/api/problem";
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

  if (status === "authenticated" && !mustChangePassword) return children;

  if (status === "loading" && error) {
    return (
      <main className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-4 px-4 py-12">
        <Alert variant="destructive">
          <AlertDescription>{messageFor(error)}</AlertDescription>
        </Alert>
        <Button variant="outline" onClick={refetch}>
          Tentar de novo
        </Button>
      </main>
    );
  }

  return (
    <main className="mx-auto flex w-full max-w-md flex-1 flex-col justify-center gap-3 px-4 py-12">
      <div role="status" aria-live="polite" className="flex flex-col gap-3">
        <Skeleton className="h-6 w-1/2" />
        <Skeleton className="h-4 w-full" />
        <Skeleton className="h-4 w-3/4" />
        <span className="sr-only">Carregando…</span>
      </div>
    </main>
  );
}
