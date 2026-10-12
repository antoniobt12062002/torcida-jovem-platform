import type { Metadata } from "next";
import Link from "next/link";

import { AccessLayout } from "@/components/app/access-layout";
import { buttonVariants } from "@/components/ui/button";
import { routes } from "@/lib/session/routes";

import { ErrorHeading } from "./error-heading";

export const metadata: Metadata = {
  title: "Página não encontrada",
};

// Endereço inexistente (DSS-04 AC3 a AC5), no layout de acesso: sem sessão
// garantida, a página não depende do shell.
export default function NotFound() {
  return (
    <AccessLayout>
      <div className="flex flex-col gap-6">
        <div className="flex flex-col gap-4">
          <ErrorHeading>Página não encontrada</ErrorHeading>
          <p className="text-sm text-pretty text-muted-foreground">
            O endereço que você abriu não existe ou mudou de lugar.
          </p>
        </div>
        <Link href={routes.inicio} className={buttonVariants({ size: "lg", className: "w-full" })}>
          Ir para o início
        </Link>
      </div>
    </AccessLayout>
  );
}
