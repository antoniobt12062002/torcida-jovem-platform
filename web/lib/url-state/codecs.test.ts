import { describe, expect, it } from "vitest";

import { codecs } from "./codecs";

// DSP-05 AC3 e AC4: o valor volta da query string; ausente ou inválido cai no
// padrão, sem erro. O padrão é serializado como `null` (parâmetro removido).

describe("codecs.string", () => {
  it("lê o texto como está e usa vazio quando o parâmetro falta", () => {
    const codec = codecs.string();
    expect(codec.parse("camisa 2026")).toBe("camisa 2026");
    expect(codec.parse(null)).toBe("");
  });

  it("serializa o padrão como null e o resto como texto", () => {
    const codec = codecs.string();
    expect(codec.serialize("")).toBeNull();
    expect(codec.serialize("camisa")).toBe("camisa");
  });

  it("aceita um padrão próprio", () => {
    const codec = codecs.string("data");
    expect(codec.parse(null)).toBe("data");
    expect(codec.serialize("data")).toBeNull();
    expect(codec.serialize("valor")).toBe("valor");
  });
});

describe("codecs.oneOf", () => {
  const tipos = ["RECEITA", "DESPESA"] as const;

  it("aceita só os valores da lista", () => {
    const codec = codecs.oneOf(tipos, "");
    expect(codec.parse("RECEITA")).toBe("RECEITA");
    expect(codec.parse("DESPESA")).toBe("DESPESA");
  });

  it("valor fora da lista, vazio ou ausente cai no padrão", () => {
    const codec = codecs.oneOf(tipos, "");
    expect(codec.parse("receita")).toBe("");
    expect(codec.parse("OUTRO")).toBe("");
    expect(codec.parse("")).toBe("");
    expect(codec.parse(null)).toBe("");
  });

  it("o padrão pode ser um dos valores (aba inicial)", () => {
    const codec = codecs.oneOf(["dados", "historico"] as const, "dados");
    expect(codec.parse(null)).toBe("dados");
    expect(codec.parse("x")).toBe("dados");
    expect(codec.parse("historico")).toBe("historico");
    expect(codec.serialize("dados")).toBeNull();
    expect(codec.serialize("historico")).toBe("historico");
  });

  it("serializa o padrão como null e um valor da lista como ele mesmo", () => {
    const codec = codecs.oneOf(tipos, "");
    expect(codec.serialize("")).toBeNull();
    expect(codec.serialize("DESPESA")).toBe("DESPESA");
  });

  it("não escreve na URL um valor fora da lista", () => {
    const codec = codecs.oneOf(tipos, "");
    expect(codec.serialize("OUTRO" as "RECEITA")).toBeNull();
  });
});

describe("codecs.isoDate", () => {
  const codec = codecs.isoDate();

  it("aceita data AAAA-MM-DD que existe no calendário", () => {
    expect(codec.parse("2026-10-11")).toBe("2026-10-11");
    expect(codec.parse("2024-02-29")).toBe("2024-02-29");
  });

  it.each([
    ["ausente", null],
    ["vazia", ""],
    ["texto", "ontem"],
    ["formato brasileiro", "11/10/2026"],
    ["sem zeros", "2026-1-5"],
    ["com hora", "2026-10-11T10:00:00Z"],
    ["mês 13", "2026-13-01"],
    ["dia 0", "2026-10-00"],
    ["30 de fevereiro", "2026-02-30"],
    ["29 de fevereiro em ano comum", "2026-02-29"],
    ["31 de abril", "2026-04-31"],
    ["espaço sobrando", " 2026-10-11"],
  ])("data inválida (%s) cai no padrão vazio", (_caso, raw) => {
    expect(codec.parse(raw)).toBe("");
  });

  it("serializa vazio e data inválida como null, e data válida como ela mesma", () => {
    expect(codec.serialize("")).toBeNull();
    expect(codec.serialize("2026-02-30")).toBeNull();
    expect(codec.serialize("2026-10-11")).toBe("2026-10-11");
  });
});
