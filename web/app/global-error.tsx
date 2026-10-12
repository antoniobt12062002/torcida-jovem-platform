"use client";

import "./globals.css";

import { Barlow_Condensed, Geist } from "next/font/google";

import { AccessLayout } from "@/components/app/access-layout";
import { Button } from "@/components/ui/button";

import { ErrorHeading } from "./error-heading";

// Exceção no layout raiz (DSS-04 AC2, AC4, AC5). Substitui o layout raiz, por
// isso traz o próprio `html` e `body`, a folha global e as fontes, como manda
// a convenção do Next 16 (error.md, "Global Error"). O erro recebido não é
// exibido. `retry` é a função de nova tentativa da convenção.

const geistSans = Geist({ variable: "--font-geist-sans", subsets: ["latin"] });

const barlowCondensed = Barlow_Condensed({
  variable: "--font-barlow-condensed",
  weight: ["600", "700"],
  subsets: ["latin"],
  display: "swap",
});

type GlobalErrorProps = {
  error: Error & { digest?: string };
  retry: () => void;
};

export default function GlobalError({ retry }: GlobalErrorProps) {
  return (
    <html lang="pt-BR" className={`${geistSans.variable} ${barlowCondensed.variable} h-full antialiased`}>
      <body className="flex min-h-full flex-col">
        <title>Algo deu errado | TJ Platform</title>
        <AccessLayout>
          <div className="flex flex-col gap-6">
            <div className="flex flex-col gap-4">
              <ErrorHeading>Algo deu errado</ErrorHeading>
              <p className="text-sm text-pretty text-muted-foreground">
                Não foi possível abrir o sistema. Tente de novo em instantes.
              </p>
            </div>
            <Button size="lg" className="w-full" onClick={() => retry()}>
              Tentar de novo
            </Button>
          </div>
        </AccessLayout>
      </body>
    </html>
  );
}
