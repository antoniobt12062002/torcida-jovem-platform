import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { useState } from "react";
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

// DSP-03 AC2: a confirmação nomeia o registro e descreve a consequência.
describe("registro e consequência", () => {
  it("mostra o registro afetado e a consequência, ligados ao diálogo como descrição", async () => {
    render(
      <ConfirmDialog
        open
        title="Cancelar lançamento?"
        subject="Receita fb6e3d8f"
        consequence="O lançamento deixa de contar no saldo. Esta ação não pode ser desfeita."
        confirmLabel="Cancelar lançamento"
        cancelLabel="Voltar"
        onConfirm={() => {}}
      />,
    );
    const dialog = await screen.findByRole("alertdialog", { name: "Cancelar lançamento?" });
    expect(within(dialog).getByText("Receita fb6e3d8f")).toBeTruthy();
    expect(within(dialog).getByText(/deixa de contar no saldo/)).toBeTruthy();
    const description = document.getElementById(dialog.getAttribute("aria-describedby") ?? "");
    expect(description?.textContent).toContain("Receita fb6e3d8f");
    expect(description?.textContent).toContain("O lançamento deixa de contar no saldo.");
  });
});

// DSP-03 AC5: a falha aparece dentro da confirmação, que continua aberta.
describe("falha na ação confirmada", () => {
  function Failing({ onConfirm }: { onConfirm: () => Promise<unknown> }) {
    const [error, setError] = useState<string>();
    return (
      <ConfirmDialog
        open
        title="Desativar conta?"
        confirmLabel="Desativar conta"
        error={error}
        onConfirm={async () => {
          try {
            await onConfirm();
          } catch (cause) {
            setError("A conta tem lançamentos em aberto.");
            throw cause;
          }
        }}
      />
    );
  }

  it("sem erro, não há alerta no diálogo", async () => {
    render(<ConfirmDialog open title="Desativar conta?" confirmLabel="Desativar conta" onConfirm={() => {}} />);
    const dialog = await screen.findByRole("alertdialog");
    expect(within(dialog).queryByRole("alert")).toBeNull();
  });

  it("mantém o diálogo aberto e mostra o erro dentro dele", async () => {
    const onConfirm = vi.fn().mockRejectedValue(new Error("falhou"));
    render(<Failing onConfirm={onConfirm} />);
    fireEvent.click(await screen.findByRole("button", { name: "Desativar conta" }));
    const alert = await within(screen.getByRole("alertdialog")).findByRole("alert");
    expect(alert.textContent).toContain("A conta tem lançamentos em aberto.");
    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(screen.getByRole("alertdialog")).toBeTruthy();
    // Pode tentar de novo: confirmar volta a ficar habilitado.
    await waitFor(() =>
      expect((screen.getByRole("button", { name: "Desativar conta" }) as HTMLButtonElement).disabled).toBe(false),
    );
  });
});

// DSC-02 AC4: com a ação em andamento, confirmar fica desabilitado, expõe
// `aria-busy` e mostra um rótulo de progresso terminado em "…".
describe("ação em andamento", () => {
  it("com pending, confirmar fica ocupado e mostra Confirmando…", async () => {
    render(<ConfirmDialog open pending title="Desativar conta?" confirmLabel="Desativar conta" onConfirm={() => {}} />);
    const dialog = await screen.findByRole("alertdialog");
    const button = within(dialog).getByRole("button", { name: "Confirmando…" }) as HTMLButtonElement;
    expect(button.getAttribute("aria-busy")).toBe("true");
    expect(button.disabled).toBe(true);
    expect(within(dialog).queryByRole("button", { name: "Desativar conta" })).toBeNull();
  });

  it("aceita um rótulo de progresso próprio", async () => {
    render(
      <ConfirmDialog
        open
        pending
        title="Cancelar lançamento?"
        confirmLabel="Cancelar lançamento"
        pendingLabel="Cancelando lançamento…"
        onConfirm={() => {}}
      />,
    );
    const button = await screen.findByRole("button", { name: "Cancelando lançamento…" });
    expect(button.getAttribute("aria-busy")).toBe("true");
  });

  it("enquanto a Promise de onConfirm não termina, confirmar fica ocupado; depois volta ao rótulo", async () => {
    let finish: (value?: unknown) => void = () => {};
    const onConfirm = vi.fn(() => new Promise((_resolve, reject) => (finish = () => reject(new Error("falhou")))));
    render(<ConfirmDialog open title="Desativar conta?" confirmLabel="Desativar conta" onConfirm={onConfirm} />);
    fireEvent.click(await screen.findByRole("button", { name: "Desativar conta" }));
    const busy = await screen.findByRole("button", { name: "Confirmando…" });
    expect(busy.getAttribute("aria-busy")).toBe("true");
    expect((busy as HTMLButtonElement).disabled).toBe(true);
    finish();
    const idle = await screen.findByRole("button", { name: "Desativar conta" });
    expect(idle.hasAttribute("aria-busy")).toBe(false);
  });

  it("sem ação em andamento, confirmar não expõe aria-busy", async () => {
    render(<ConfirmDialog open title="Desativar conta?" confirmLabel="Desativar conta" onConfirm={() => {}} />);
    const button = await screen.findByRole("button", { name: "Desativar conta" });
    expect(button.hasAttribute("aria-busy")).toBe(false);
  });
});

// DSP-03 AC3: o rótulo de confirmar é sempre informado por quem chama.
// Conferido pelo `tsc` (gate Web), sem execução.
export function recusadoPeloCompilador() {
  // @ts-expect-error `confirmLabel` é obrigatório: não existe "Confirmar" genérico.
  return <ConfirmDialog title="Desativar conta?" onConfirm={() => {}} />;
}

// DSC-02 AC4: o rótulo de progresso termina em "…". Conferido pelo `tsc`.
export function rotuloDeProgressoSemReticencias() {
  return (
    // @ts-expect-error `pendingLabel` precisa terminar em "…".
    <ConfirmDialog title="Desativar conta?" confirmLabel="Desativar conta" pendingLabel="Desativando" onConfirm={() => {}} />
  );
}
