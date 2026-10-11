import { ApiError, type ErrorCatalog } from "@/lib/api/problem";
import { formatRetryAfter } from "@/lib/session/retry-after";

// Catálogo dos erros de senha da feature de acesso (ACS-01 AC5-6, ACS-03
// AC4-5). A API devolve o `code` no corpo, sem `errors[]`: o campo de cada
// código é decidido aqui.

/** Campos de senha dos formulários de troca e de redefinição. */
export type PasswordField = "current_password" | "new_password";

const MIN_LENGTH_ADMIN = 10;
const MIN_LENGTH_DEFAULT = 8;

/**
 * Tamanho mínimo exibido como dica: 10 para quem tem vínculo administrativo,
 * 8 para os demais. A API continua decidindo.
 */
export function minPasswordLength(adminMembership: unknown): number {
  return adminMembership ? MIN_LENGTH_ADMIN : MIN_LENGTH_DEFAULT;
}

export function minLengthHint(min: number): string {
  return `Mínimo de ${min} caracteres.`;
}

/** Dica de quem não tem sessão (redefinição pelo link): não se sabe o vínculo. */
export const MIN_LENGTH_HINT_UNKNOWN = `Mínimo de ${MIN_LENGTH_DEFAULT} caracteres (${MIN_LENGTH_ADMIN} para quem tem vínculo administrativo).`;

const MESSAGE_TOO_SHORT = "A senha é curta demais.";

const FIELD_ERRORS: Readonly<Record<string, { field: PasswordField; message: string }>> = {
  invalid_current_password: { field: "current_password", message: "A senha atual está incorreta." },
  password_unchanged: {
    field: "new_password",
    message: "A nova senha precisa ser diferente da atual.",
  },
  password_too_short: { field: "new_password", message: MESSAGE_TOO_SHORT },
  password_too_long: { field: "new_password", message: "A senha é longa demais." },
  password_compromised: {
    field: "new_password",
    message: "Esta senha aparece em vazamentos conhecidos. Escolha outra.",
  },
};

/**
 * Campo e mensagem de um erro de senha da API, ou null quando o código não é
 * de campo. `minHint` completa a mensagem de senha curta.
 */
export function passwordFieldError(
  code: string,
  minHint: string,
): { field: PasswordField; message: string } | null {
  const entry = Object.hasOwn(FIELD_ERRORS, code) ? FIELD_ERRORS[code] : undefined;
  if (!entry) return null;
  if (code === "password_too_short") return { ...entry, message: `${MESSAGE_TOO_SHORT} ${minHint}` };
  return entry;
}

export const MESSAGE_CONFIRM_MISMATCH = "A confirmação não confere com a nova senha.";

export const MESSAGE_INVALID_RESET_TOKEN = "O link expirou ou já foi usado.";

export function blockedMessage(retryAfter: string | null): string {
  const wait = formatRetryAfter(retryAfter);
  return wait ? `Muitas tentativas. Tente de novo em ${wait}.` : "Muitas tentativas. Tente de novo mais tarde.";
}

/** Mensagens de formulário (sem campo) da troca de senha. */
export function changePasswordCatalog(retryAfter: string | null): ErrorCatalog {
  return { password_change_blocked: blockedMessage(retryAfter) };
}

/** Erro da API com o cabeçalho Retry-After da resposta. */
export class PasswordApiError extends ApiError {
  readonly retryAfter: string | null;

  constructor(error: ApiError, retryAfter: string | null) {
    super({
      status: error.status,
      code: error.code,
      title: error.title,
      detail: error.detail,
      errors: error.errors,
      cause: error.cause,
    });
    this.name = "PasswordApiError";
    this.retryAfter = retryAfter;
  }
}
