import { describe, expect, it } from "vitest";

import { loginPath, routes, safeNext } from "./routes";

describe("routes", () => {
  it("tem os caminhos usados pela sessão e pelas rotas públicas", () => {
    expect(routes).toEqual({
      entrar: "/entrar",
      contaSenha: "/conta/senha",
      recuperarAcesso: "/recuperar-acesso",
      redefinirSenha: "/redefinir-senha",
      inicio: "/inicio",
      semAcesso: "/sem-acesso",
    });
  });
});

// FND-03 AC1 e caso de borda: `next` só é aceito se for caminho interno.
describe("safeNext", () => {
  it.each([
    "/financeiro/contas",
    "/estoque/produtos?ativo=true",
    "/admin/usuarios#lista",
    "/",
  ])("aceita o caminho interno %s", (next) => {
    expect(safeNext(next)).toBe(next);
  });

  it.each([
    ["outra origem sem esquema", "//exemplo.com"],
    ["outra origem com barras extras", "///exemplo.com"],
    ["URL absoluta https", "https://exemplo.com/inicio"],
    ["URL absoluta http", "http://localhost:3000/inicio"],
    ["esquema javascript", "javascript:alert(1)"],
    ["barra invertida no início", "/\\exemplo.com"],
    ["barra invertida no meio", "/financeiro\\..\\x"],
    ["caminho relativo", "financeiro/contas"],
    ["caminho relativo com ponto", "./inicio"],
    ["texto vazio", ""],
    ["só espaços", "   "],
    ["espaço antes da barra", " /inicio"],
    ["caractere de controle", "/\t/exemplo.com"],
    ["quebra de linha", "/\n/exemplo.com"],
  ])("recusa %s e usa a página inicial", (_caso, next) => {
    expect(safeNext(next)).toBe("/inicio");
  });

  it("usa a página inicial sem next", () => {
    expect(safeNext(null)).toBe("/inicio");
    expect(safeNext(undefined)).toBe("/inicio");
  });

  it("usa o primeiro valor quando next vem repetido", () => {
    expect(safeNext(["/financeiro/contas", "//exemplo.com"])).toBe("/financeiro/contas");
    expect(safeNext(["//exemplo.com", "/financeiro/contas"])).toBe("/inicio");
  });
});

describe("loginPath", () => {
  function parse(path: string) {
    const url = new URL(path, "http://localhost:3000");
    return { pathname: url.pathname, next: url.searchParams.get("next"), url };
  }

  it("leva a /entrar sem parâmetros", () => {
    expect(loginPath()).toBe("/entrar");
  });

  it("leva o caminho pedido em next, codificado", () => {
    const { pathname, next } = parse(loginPath({ next: "/financeiro/contas?x=1&y=2" }));
    expect(pathname).toBe("/entrar");
    expect(next).toBe("/financeiro/contas?x=1&y=2");
  });

  it("não leva next que não seria aceito depois", () => {
    expect(parse(loginPath({ next: "//exemplo.com" })).next).toBeNull();
  });
});
