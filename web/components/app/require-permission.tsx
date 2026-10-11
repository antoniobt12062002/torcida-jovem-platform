"use client";

import Link from "next/link";
import type { ReactNode } from "react";

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

/** Conteúdo da página "Sem acesso" (o mesmo de /sem-acesso). */
export function NoAccess() {
  return (
    <section className="mx-auto flex w-full max-w-md flex-col justify-center gap-3 px-4 py-12">
      <h1 className="text-xl font-semibold">Sem acesso</h1>
      <p className="text-sm text-muted-foreground">
        Você não tem permissão para acessar esta área. Se precisar dela, fale com a administração.
      </p>
      <Link href={routes.inicio} className="text-sm underline underline-offset-4 hover:text-foreground">
        Voltar ao início
      </Link>
    </section>
  );
}
