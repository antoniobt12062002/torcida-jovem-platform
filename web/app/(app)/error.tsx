"use client";

import Link from "next/link";

import { PageContainer } from "@/components/app/page-container";
import { Button, buttonVariants } from "@/components/ui/button";
import { routes } from "@/lib/session/routes";

import { ErrorHeading } from "../error-heading";

// Exceção de renderização numa rota autenticada (DSS-04 AC1, AC4, AC5). O Next
// envolve esta página com o layout do grupo, então o shell continua na tela.
// O erro recebido não é exibido: nem mensagem, nem pilha, nem `digest`.
// `retry` é a função de nova tentativa da convenção do Next 16 (error.md).

type AppErrorProps = {
  error: Error & { digest?: string };
  retry: () => void;
};

export default function AppError({ retry }: AppErrorProps) {
  return (
    <PageContainer width="narrow">
      <div className="flex flex-col gap-4">
        <ErrorHeading documentTitle="Algo deu errado | TJ Platform">Algo deu errado</ErrorHeading>
        <p className="text-sm text-pretty text-muted-foreground">
          Não foi possível mostrar esta página. Tente de novo; se o problema continuar, volte ao início.
        </p>
      </div>
      <div className="flex flex-col gap-2 *:w-full md:flex-row md:*:w-auto">
        <Button onClick={() => retry()}>Tentar de novo</Button>
        <Link href={routes.inicio} className={buttonVariants({ variant: "secondary" })}>
          Voltar ao início
        </Link>
      </div>
    </PageContainer>
  );
}
