// Token de recuperação do fragmento do link do e-mail (`#token=…`, ACS-03 AC1).
// O fragmento não vai ao servidor; quem chama remove-o da barra de endereço.
// Leitura manual, sem URLSearchParams, para não trocar "+" por espaço.

export function readResetToken(hash: string): string | null {
  const fragment = hash.startsWith("#") ? hash.slice(1) : hash;
  for (const part of fragment.split("&")) {
    if (!part.startsWith("token=")) continue;
    let token: string;
    try {
      token = decodeURIComponent(part.slice("token=".length));
    } catch {
      return null;
    }
    return token.trim() === "" ? null : token;
  }
  return null;
}
