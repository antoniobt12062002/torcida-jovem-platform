import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";

import { activeStatus, StatusBadge, type Status } from "./status-badge";

describe("StatusBadge", () => {
  it.each<[Status, string]>([
    ["CRIADA", "Criada"],
    ["RECEBIDA", "Recebida"],
    ["PAGA", "Paga"],
    ["CANCELADA", "Cancelada"],
    ["ativo", "Ativo"],
    ["inativo", "Inativo"],
  ])("mostra %s como %s", (status, label) => {
    render(<StatusBadge status={status} />);
    expect(screen.getByText(label)).toBeTruthy();
  });

  it("activeStatus converte o booleano de ativo", () => {
    expect(activeStatus(true)).toBe("ativo");
    expect(activeStatus(false)).toBe("inativo");
  });
});
