"use client";

import Link from "next/link";
import { type FormEvent, useState } from "react";

import { FormLayout } from "@/components/app/form-layout";
import { Button, buttonVariants } from "@/components/ui/button";
import { Field, FieldError, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { toast } from "@/components/ui/toast";
import { PageTitle } from "@/components/ui/typography";

// Formulário de login de exemplo: valida os campos no navegador e não envia
// nada. O e-mail e a senha digitados não saem da página.

export function ExampleLoginForm() {
  const [email, setEmail] = useState("");
  const [senha, setSenha] = useState("");
  const [submitted, setSubmitted] = useState(false);
  const emailMissing = email.trim() === "";
  const senhaMissing = senha === "";

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSubmitted(true);
    if (emailMissing || senhaMissing) return;
    toast.show("Exemplo: no espécime o login não é enviado.");
  }

  return (
    <div className="flex flex-col gap-6">
      <div className="flex flex-col gap-2">
        <PageTitle>Entrar</PageTitle>
        <p className="text-sm text-muted-foreground">Use o e-mail e a senha da sua conta.</p>
      </div>
      <FormLayout
        noValidate
        onSubmit={handleSubmit}
        footer={
          <Button type="submit" size="lg" className="w-full md:w-full">
            Entrar
          </Button>
        }
      >
        <Field invalid={submitted && emailMissing}>
          <FieldLabel>E-mail</FieldLabel>
          <Input
            type="email"
            autoComplete="off"
            spellCheck={false}
            autoCapitalize="none"
            placeholder="pessoa@exemplo.com…"
            value={email}
            onChange={(event) => setEmail(event.target.value)}
          />
          {submitted && emailMissing ? <FieldError>Informe o e-mail.</FieldError> : null}
        </Field>
        <Field invalid={submitted && senhaMissing}>
          <FieldLabel>Senha</FieldLabel>
          <Input type="password" autoComplete="off" value={senha} onChange={(event) => setSenha(event.target.value)} />
          {submitted && senhaMissing ? <FieldError>Informe a senha.</FieldError> : null}
        </Field>
      </FormLayout>
      <div>
        <Link href="/design-system#acesso" className={buttonVariants({ variant: "link", size: "sm" })}>
          Voltar ao espécime
        </Link>
      </div>
    </div>
  );
}
