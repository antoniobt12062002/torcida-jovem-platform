"use client";

import { type ComponentProps, useEffect, useRef } from "react";

import { Alert, AlertDescription } from "@/components/ui/alert";
import { cn } from "@/lib/utils";

// Erro de um formulário ou de uma confirmação que não pertence a nenhum campo
// (DSP-04 AC4, DSP-03 AC5). Recebe o foco ao surgir: durante o envio os
// controles ficam desabilitados e o foco cairia no body, fora do diálogo.

export type FormErrorAlertProps = Omit<ComponentProps<typeof Alert>, "variant" | "tabIndex" | "ref">;

export function FormErrorAlert({ className, children, ...props }: FormErrorAlertProps) {
  const ref = useRef<HTMLDivElement>(null);

  // Quem usa monta o alerta só quando há erro: o foco vai a ele uma vez.
  useEffect(() => {
    ref.current?.focus();
  }, []);

  return (
    <Alert
      ref={ref}
      variant="error"
      tabIndex={-1}
      className={cn(
        "outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-ring-offset",
        className,
      )}
      {...props}
    >
      <AlertDescription>{children}</AlertDescription>
    </Alert>
  );
}
