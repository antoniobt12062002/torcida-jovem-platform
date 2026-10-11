import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";

import { Button } from "@/components/ui/button";

import { ConfirmDialog } from "./confirm-dialog";

// FND-05 AC6: confirmação explícita antes de chamar a API; com motivo
// obrigatório, confirmar fica desabilitado com motivo vazio ou só espaços.

function confirmButton() {
  return screen.getByRole("button", { name: "Cancelar lançamento" }) as HTMLButtonElement;
}

describe("confirmação explícita", () => {
  it("só chama onConfirm ao confirmar; cancelar não chama", async () => {
    const onConfirm = vi.fn();
    render(
      <ConfirmDialog
        trigger={<Button>Excluir</Button>}
        title="Desativar conta?"
        description="A conta deixa de aceitar lançamentos."
        confirmLabel="Desativar"
        onConfirm={onConfirm}
      />,
    );
    expect(onConfirm).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Excluir" }));
    await screen.findByRole("alertdialog", { name: "Desativar conta?" });
    expect(onConfirm).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Cancelar" }));
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
    expect(onConfirm).not.toHaveBeenCalled();

    fireEvent.click(screen.getByRole("button", { name: "Excluir" }));
    await screen.findByRole("alertdialog");
    fireEvent.click(screen.getByRole("button", { name: "Desativar" }));
    await waitFor(() => expect(onConfirm).toHaveBeenCalledTimes(1));
    expect(onConfirm).toHaveBeenCalledWith(undefined);
    await waitFor(() => expect(screen.queryByRole("alertdialog")).toBeNull());
  });

  it("se a ação falhar, o diálogo continua aberto", async () => {
    const onConfirm = vi.fn().mockRejectedValue(new Error("falhou"));
    render(<ConfirmDialog open title="Desativar conta?" confirmLabel="Desativar" onConfirm={onConfirm} />);
    fireEvent.click(await screen.findByRole("button", { name: "Desativar" }));
    await waitFor(() => expect(onConfirm).toHaveBeenCalledTimes(1));
    await waitFor(() =>
      expect((screen.getByRole("button", { name: "Desativar" }) as HTMLButtonElement).disabled).toBe(false),
    );
    expect(screen.getByRole("alertdialog")).toBeTruthy();
  });
});

describe("motivo obrigatório", () => {
  async function renderWithReason(onConfirm = vi.fn()) {
    const onOpenChange = vi.fn();
    render(
      <ConfirmDialog
        open
        onOpenChange={onOpenChange}
        title="Cancelar lançamento?"
        confirmLabel="Cancelar lançamento"
        cancelLabel="Voltar"
        reason="required"
        onConfirm={onConfirm}
      />,
    );
    await screen.findByRole("alertdialog");
    return { onConfirm, onOpenChange, reason: screen.getByLabelText("Motivo") as HTMLTextAreaElement };
  }

  it("desabilitado com motivo vazio", async () => {
    const { onConfirm } = await renderWithReason();
    expect(confirmButton().disabled).toBe(true);
    fireEvent.click(confirmButton());
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("desabilitado com motivo só de espaços", async () => {
    const { onConfirm, reason } = await renderWithReason();
    fireEvent.change(reason, { target: { value: "  \n\t " } });
    expect(confirmButton().disabled).toBe(true);
    fireEvent.click(confirmButton());
    expect(onConfirm).not.toHaveBeenCalled();
  });

  it("com motivo, habilita e entrega o motivo aparado; fecha ao concluir", async () => {
    const { onConfirm, onOpenChange, reason } = await renderWithReason(vi.fn().mockResolvedValue(undefined));
    fireEvent.change(reason, { target: { value: "  lançado em duplicidade  " } });
    expect(confirmButton().disabled).toBe(false);
    fireEvent.click(confirmButton());
    await waitFor(() => expect(onConfirm).toHaveBeenCalledWith("lançado em duplicidade"));
    await waitFor(() => expect(onOpenChange).toHaveBeenCalledWith(false));
  });
});
