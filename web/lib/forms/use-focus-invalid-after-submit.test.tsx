import { cleanup, fireEvent, render, screen, waitFor } from "@testing-library/react";
import { useForm } from "react-hook-form";
import { afterEach, describe, expect, it } from "vitest";

import { useFocusInvalidAfterSubmit } from "./use-focus-invalid-after-submit";

afterEach(cleanup);

type Values = { nome: string; codigo: string };

// O envio desabilita os campos, como nos diálogos; a "API" devolve erros de
// campo fora da ordem do formulário.
function Harness({ errors }: { errors: Array<keyof Values | "root"> }) {
  const form = useForm<Values>({ defaultValues: { nome: "", codigo: "" } });
  const {
    register,
    handleSubmit,
    setError,
    formState: { isSubmitting },
  } = form;
  useFocusInvalidAfterSubmit(form, ["nome", "codigo"]);
  return (
    <form
      noValidate
      onSubmit={(event) =>
        void handleSubmit(async () => {
          await new Promise((resolve) => setTimeout(resolve, 5));
          for (const name of errors) setError(name, { message: `erro em ${name}` });
        })(event)
      }
    >
      <input aria-label="Nome" disabled={isSubmitting} {...register("nome")} />
      <input aria-label="Código" disabled={isSubmitting} {...register("codigo")} />
      <button type="submit">Enviar</button>
    </form>
  );
}

describe("useFocusInvalidAfterSubmit", () => {
  it("ao fim do envio, foca o primeiro campo inválido na ordem do formulário", async () => {
    render(<Harness errors={["codigo", "nome"]} />);
    const button = screen.getByRole("button", { name: "Enviar" });
    button.focus();
    fireEvent.click(button);
    await waitFor(() => expect(document.activeElement).toBe(screen.getByLabelText("Nome")));
  });

  it("com erro sem campo (root), não move o foco: o alerta cuida dele", async () => {
    render(<Harness errors={["root", "codigo"]} />);
    const button = screen.getByRole("button", { name: "Enviar" });
    button.focus();
    fireEvent.click(button);
    await new Promise((resolve) => setTimeout(resolve, 30));
    expect(document.activeElement).toBe(button);
  });

  it("sem erro, não move o foco", async () => {
    render(<Harness errors={[]} />);
    const button = screen.getByRole("button", { name: "Enviar" });
    button.focus();
    fireEvent.click(button);
    await new Promise((resolve) => setTimeout(resolve, 30));
    expect(document.activeElement).toBe(button);
  });
});
