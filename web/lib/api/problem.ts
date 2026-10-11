// Erros da API (application/problem+json, RFC 9457) como um tipo só para as
// telas: FND-01 AC4 e FND-06.

export type FieldError = { field: string; code: string };

/** Catálogo de mensagens de uma tela: `code` da API → texto em português. */
export type ErrorCatalog = Readonly<Record<string, string>>;

export const NETWORK_ERROR_CODE = "network_error";

export const MESSAGE_PAYLOAD_TOO_LARGE = "O conteúdo enviado é grande demais.";
export const MESSAGE_UNAVAILABLE =
  "O serviço está indisponível no momento. Tente de novo em instantes.";

type ApiErrorInit = {
  status: number;
  code: string;
  title: string;
  detail?: string;
  errors?: FieldError[];
  cause?: unknown;
};

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly title: string;
  readonly detail?: string;
  readonly errors: FieldError[];

  constructor({ status, code, title, detail, errors = [], cause }: ApiErrorInit) {
    super(`${status} ${code}`, cause === undefined ? undefined : { cause });
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.title = title;
    this.detail = detail;
    this.errors = errors;
  }
}

export function isApiError(value: unknown): value is ApiError {
  return value instanceof ApiError;
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function fieldErrors(value: unknown): FieldError[] {
  if (!Array.isArray(value)) return [];
  return value
    .filter(isRecord)
    .filter((e) => typeof e.field === "string" && typeof e.code === "string")
    .map((e) => ({ field: e.field as string, code: e.code as string }));
}

/** Converte uma resposta de erro (com o corpo já lido) em ApiError. */
export function toApiError(response: Response, body: unknown): ApiError {
  const problem = isRecord(body) ? body : {};
  return new ApiError({
    status: response.status,
    code: typeof problem.code === "string" ? problem.code : `http_${response.status}`,
    title: typeof problem.title === "string" ? problem.title : response.statusText,
    detail: typeof problem.detail === "string" ? problem.detail : undefined,
    errors: fieldErrors(problem.errors),
  });
}

/** Falha antes de haver resposta (rede, DNS, servidor Next fora do ar). */
export function networkError(cause?: unknown): ApiError {
  return new ApiError({ status: 0, code: NETWORK_ERROR_CODE, title: "Network error", cause });
}

/** Indisponibilidade da API: a tela oferece tentar de novo (FND-06 AC5). */
export function isUnavailable(error: ApiError): boolean {
  return error.status === 503 || error.code === NETWORK_ERROR_CODE;
}

/**
 * Mensagem para a pessoa: a do catálogo da tela para o `code`, senão uma
 * genérica. Nunca expõe `title`, `detail` nem o `code` (FND-06 AC2-5).
 */
export function messageFor(error: ApiError, catalog: ErrorCatalog = {}): string {
  const known = Object.hasOwn(catalog, error.code) ? catalog[error.code] : undefined;
  if (known) return known;
  if (error.status === 413) return MESSAGE_PAYLOAD_TOO_LARGE;
  if (isUnavailable(error)) return MESSAGE_UNAVAILABLE;
  return `Não foi possível concluir a operação (erro HTTP ${error.status}).`;
}
