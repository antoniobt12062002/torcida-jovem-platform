// Caminhos usados pela sessão e pelas rotas públicas e autenticadas da
// fundação. As páginas de cada caminho pertencem às suas unidades.

export const routes = {
  entrar: "/entrar",
  contaSenha: "/conta/senha",
  recuperarAcesso: "/recuperar-acesso",
  redefinirSenha: "/redefinir-senha",
  inicio: "/inicio",
  semAcesso: "/sem-acesso",
} as const;

/** Parâmetro de /entrar que pede o aviso de sessão encerrada. */
export const SESSION_ENDED_PARAM = "sessao";
export const SESSION_ENDED_VALUE = "encerrada";

const PROBE_ORIGIN = "http://origem.invalid";
/** Caracteres de controle (inclui tab e quebras de linha), espaço e DEL. */
function hasUnsafeChar(value: string): boolean {
  for (let i = 0; i < value.length; i++) {
    const code = value.charCodeAt(i);
    if (code <= 0x20 || code === 0x7f) return true;
  }
  return false;
}

function isInternalPath(value: string): boolean {
  if (!value.startsWith("/")) return false;
  // "//x" e "/\x" são lidos pelo navegador como outra origem.
  if (value.startsWith("//") || value.includes("\\")) return false;
  if (hasUnsafeChar(value)) return false;
  try {
    return new URL(value, PROBE_ORIGIN).origin === PROBE_ORIGIN;
  } catch {
    return false;
  }
}

/**
 * Destino depois do login: só um caminho interno começando com `/`; qualquer
 * outra coisa (outra origem, esquema, caminho relativo) vira a página inicial.
 */
export function safeNext(next: string | string[] | null | undefined): string {
  const value = Array.isArray(next) ? next[0] : next;
  return typeof value === "string" && isInternalPath(value) ? value : routes.inicio;
}

/** Caminho de /entrar, com o destino pedido e o aviso de sessão encerrada. */
export function loginPath({ next, sessionEnded = false }: { next?: string | null; sessionEnded?: boolean } = {}): string {
  const params = new URLSearchParams();
  if (next && isInternalPath(next)) params.set("next", next);
  if (sessionEnded) params.set(SESSION_ENDED_PARAM, SESSION_ENDED_VALUE);
  const query = params.toString();
  return query ? `${routes.entrar}?${query}` : routes.entrar;
}
