import createClient, { type Middleware } from "openapi-fetch";

import type { paths as EstoquePaths } from "./estoque";
import type { paths as FinanceiroPaths } from "./financeiro";
import type { paths as IdentityPaths } from "./identity";
import { isApiError, networkError, toApiError } from "./problem";

// Cliente tipado da API (FND-01 AC2-4). O navegador chama só a própria origem;
// o rewrite do next.config.ts leva /api/v1/* até a API (ADR-010).
//
// lib/api não importa lib/session: a sessão registra aqui, de fora, de onde vem
// o token CSRF e como reagir a 401/403 (configureApiClient).

const CSRF_HEADER = "X-CSRF-Token";
const SAFE_METHODS = new Set(["GET", "HEAD", "OPTIONS"]);
const LOGIN_PATH = "/api/v1/auth/login";

/** Dados da resposta que disparou um callback. */
export type ApiFailure = { status: number; code: string | undefined; method: string; path: string };

export type ApiClientHooks = {
  /** Token CSRF da sessão atual, guardado em memória por lib/session. */
  getCsrfToken: () => string | null | undefined;
  /** 401 em qualquer chamada, exceto POST /auth/login. */
  onUnauthenticated: (failure: ApiFailure) => void;
  /** 403 com code password_change_required. */
  onPasswordChangeRequired: (failure: ApiFailure) => void;
  /** 403 com code csrf_invalid. */
  onCsrfInvalid: (failure: ApiFailure) => void;
};

const noop = () => {};
const defaultHooks: ApiClientHooks = {
  getCsrfToken: () => undefined,
  onUnauthenticated: noop,
  onPasswordChangeRequired: noop,
  onCsrfInvalid: noop,
};
let hooks: ApiClientHooks = defaultHooks;

/**
 * Registra os callbacks do cliente. Devolve a função que os desfaz (os
 * callbacks voltam a não fazer nada).
 */
export function configureApiClient(next: Partial<ApiClientHooks>): () => void {
  const registered: ApiClientHooks = { ...defaultHooks, ...next };
  hooks = registered;
  return () => {
    if (hooks === registered) hooks = defaultHooks;
  };
}

async function problemCode(response: Response): Promise<string | undefined> {
  try {
    const body: unknown = await response.clone().json();
    if (typeof body === "object" && body !== null && "code" in body) {
      return typeof body.code === "string" ? body.code : undefined;
    }
  } catch {
    // corpo ausente ou não JSON: sem code
  }
  return undefined;
}

const sessionMiddleware: Middleware = {
  onRequest({ request }) {
    if (SAFE_METHODS.has(request.method)) return;
    const token = hooks.getCsrfToken();
    if (token) request.headers.set(CSRF_HEADER, token);
    return request;
  },
  async onResponse({ request, response, schemaPath }) {
    if (response.status !== 401 && response.status !== 403) return;
    const failure: ApiFailure = {
      status: response.status,
      code: await problemCode(response),
      method: request.method,
      path: schemaPath,
    };
    if (response.status === 401) {
      if (!(request.method === "POST" && schemaPath === LOGIN_PATH)) hooks.onUnauthenticated(failure);
    } else if (failure.code === "password_change_required") {
      hooks.onPasswordChangeRequired(failure);
    } else if (failure.code === "csrf_invalid") {
      hooks.onCsrfInvalid(failure);
    }
  },
};

// O contrato declara X-CSRF-Token como parâmetro obrigatório das escritas. Quem
// o preenche é o middleware acima, então o tipo do cliente o torna opcional;
// os tipos gerados (*.d.ts) ficam intactos.
type WithoutCsrfParams<P> = P extends { header: infer H }
  ? Omit<P, "header"> & { header?: Omit<H, typeof CSRF_HEADER> }
  : P;
type WithoutCsrfOperation<O> = O extends { parameters: infer P }
  ? Omit<O, "parameters"> & { parameters: WithoutCsrfParams<P> }
  : O;
type WithoutCsrf<Paths> = {
  [Path in keyof Paths]: {
    [Method in keyof Paths[Path]]: Method extends "parameters"
      ? Paths[Path][Method]
      : WithoutCsrfOperation<Paths[Path][Method]>;
  };
};

function currentOrigin(): string {
  return typeof window === "undefined" ? "" : window.location.origin;
}

function createApiClient<Paths extends object>() {
  const client = createClient<WithoutCsrf<Paths>>({
    baseUrl: currentOrigin(),
    credentials: "same-origin",
    // Resolve o fetch global a cada chamada, não na criação do módulo.
    fetch: (request) => globalThis.fetch(request),
  });
  client.use(sessionMiddleware);
  return client;
}

/** Um cliente por contrato: os tipos de cada um vêm do seu `.d.ts`. */
export const api = {
  identity: createApiClient<IdentityPaths>(),
  financeiro: createApiClient<FinanceiroPaths>(),
  estoque: createApiClient<EstoquePaths>(),
};

type FetchResult = { data?: unknown; error?: unknown; response: Response };

/**
 * Devolve o corpo de sucesso ou lança ApiError (problem+json, outro erro HTTP
 * ou falha de rede). Feito para `queryFn` e `mutationFn` do TanStack Query.
 */
export async function unwrap<R extends FetchResult>(
  call: Promise<R>,
): Promise<NonNullable<R["data"]>> {
  let result: R;
  try {
    result = await call;
  } catch (cause) {
    if (isApiError(cause)) throw cause;
    throw networkError(cause);
  }
  if (!result.response.ok) throw toApiError(result.response, result.error);
  return result.data as NonNullable<R["data"]>;
}
