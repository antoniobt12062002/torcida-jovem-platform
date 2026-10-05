import { describe, expect, it } from "vitest";

import { authorLabel, formatDate, formatDateTime, shortId } from "./format";

// FND-05 AC7 (datas pt-BR no fuso de São Paulo) e WEB-D-014 (autoria).

describe("formatDate", () => {
  it("formata um instante em dd/mm/aaaa no fuso de São Paulo", () => {
    expect(formatDate("2026-10-04T12:00:00Z")).toBe("04/10/2026");
  });

  it("usa o dia de São Paulo, não o de UTC", () => {
    // 02:00 UTC de 04/10 ainda é 23:00 de 03/10 em São Paulo.
    expect(formatDate("2026-10-04T02:00:00Z")).toBe("03/10/2026");
  });

  it("formata uma data de calendário sem deslocar o dia", () => {
    expect(formatDate("2026-10-04")).toBe("04/10/2026");
  });
});

describe("formatDateTime", () => {
  it("formata data e hora pt-BR no fuso de São Paulo", () => {
    expect(formatDateTime("2026-10-04T12:00:00Z")).toBe("04/10/2026 09:00");
    expect(formatDateTime("2026-01-15T03:30:00Z")).toBe("15/01/2026 00:30");
  });
});

describe("autoria", () => {
  const me = "3f2a9c1e-7b4d-4e2a-9c1e-7b4d4e2a9c1e";
  const other = "a1b2c3d4-0000-4000-8000-000000000002";

  it("shortId devolve o identificador curto", () => {
    expect(shortId(other)).toBe("a1b2c3d4");
  });

  it('authorLabel devolve "você" para o id da sessão', () => {
    expect(authorLabel(me, me)).toBe("você");
  });

  it("authorLabel devolve o identificador curto para outra pessoa", () => {
    expect(authorLabel(other, me)).toBe("a1b2c3d4");
    expect(authorLabel(other, undefined)).toBe("a1b2c3d4");
  });
});
