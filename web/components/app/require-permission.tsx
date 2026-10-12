"use client";

import { LockIcon } from "lucide-react";
import Link from "next/link";
import type { ReactNode } from "react";

import { PageTitle } from "@/components/ui/typography";
import { routes } from "@/lib/session/routes";
import { useCan, useSession } from "@/lib/session/session-provider";

// Gate de permissão na interface (FND-04 AC3-4). Sem a permissão, os filhos
// não são renderizados, então nenhuma consulta da área chega a ser feita. A
// API continua sendo a autoridade.

export type RequirePermissionProps = {
  /** Permissão efetiva exigida; com uma lista, todas são exigidas. */
  permission: string | readonly string[];
  /** `page`: sem a permissão mostra "Sem acesso"; `action`: não mostra nada. */
  mode: "page" | "action";
  children: ReactNode;
};

export function RequirePermission({ permission, mode, children }: RequirePermissionProps) {
  const { status } = useSession();
  const required = typeof permission === "string" ? [permission] : permission;
  const allowed = useCan(...required);

  if (allowed) return children;
  // Enquanto a sessão não está pronta o portão das rotas já mostra o
  // carregamento; aqui não se antecipa "Sem acesso".
  if (mode === "action" || status !== "authenticated") return null;
  return <NoAccess />;
}

/**
 * Estado "Sem acesso", o mesmo em todas as áreas (DSP-02 AC5, DSP-06 AC4).
 * Ocupa o lugar do conteúdo da página e traz o único `h1` dela.
 */
export function NoAccess() {
  return (
    <section
      data-slot="no-access"
      className="mx-auto flex w-full max-w-content-narrow flex-col items-start gap-3 py-8 md:py-12"
    >
      <LockIcon aria-hidden="true" className="size-6 text-muted-foreground" />
      <PageTitle>Sem acesso</PageTitle>
      <p className="text-sm text-pretty text-muted-foreground">
        Você não tem permissão para acessar esta área. Se precisar dela, fale com a administração.
      </p>
      <Link
        href={routes.inicio}
        className="inline-flex min-h-control-sm touch-manipulation items-center rounded-sm text-sm font-medium underline underline-offset-4 outline-none hover:decoration-2 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset"
      >
        Voltar ao início
      </Link>
    </section>
  );
}
