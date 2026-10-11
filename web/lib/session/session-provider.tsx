"use client";

import { hashKey, useQuery, useQueryClient } from "@tanstack/react-query";
import { usePathname, useRouter } from "next/navigation";
import {
  createContext,
  type ReactNode,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
} from "react";

import { toast } from "@/components/ui/toast";
import { api, configureApiClient, unwrap } from "@/lib/api/client";
import type { components } from "@/lib/api/identity";
import { ApiError, isApiError, networkError, toApiError } from "@/lib/api/problem";
import { queryKeys } from "@/lib/api/query-keys";

import { loginPath, routes } from "./routes";

// Sessão da Web (FND-03, FND-04 AC3). Único dono do contexto de sessão e do
// token CSRF, guardados só em memória (no cache do TanStack Query, sob
// queryKeys.session.me): nada vai para localStorage, sessionStorage ou cookie.
// O valor do cache é o AuthContext com sessão, `null` sem sessão e ausente
// enquanto a primeira consulta não termina.

export type AuthContext = components["schemas"]["AuthContext"];

export type SessionStatus = "loading" | "authenticated" | "anonymous";

/** Erro de login, com o tempo de espera de um 429 (cabeçalho Retry-After). */
export class LoginError extends ApiError {
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
    this.name = "LoginError";
    this.retryAfter = retryAfter;
  }
}

export type SessionValue = {
  status: SessionStatus;
  context: AuthContext | null;
  /** Falha ao consultar a sessão (API indisponível, erro 5xx); sem 401. */
  error: ApiError | null;
  /** Consulta /auth/me de novo (por exemplo depois de uma falha). */
  refetch: () => void;
  /** POST /auth/login; guarda o contexto devolvido. Lança LoginError. */
  login: (email: string, password: string) => Promise<AuthContext>;
  /** POST /auth/logout; limpa o cache e leva a /entrar. */
  logout: () => Promise<void>;
  /**
   * Leva a /entrar com o caminho atual em `next`. Um único redirecionamento
   * por vez, mesmo com várias chamadas seguidas.
   */
  redirectToLogin: () => void;
};

const SessionContext = createContext<SessionValue | null>(null);

const SESSION_HASH = hashKey(queryKeys.session.me);

const MESSAGE_LOGOUT_FAILED = "Não foi possível sair. Tente de novo.";
const MESSAGE_CSRF_REFRESHED = "A sessão foi atualizada. Repita a ação, por favor.";

async function fetchSession(): Promise<AuthContext | null> {
  try {
    return await unwrap(api.identity.GET("/api/v1/auth/me"));
  } catch (error) {
    if (isApiError(error) && error.status === 401) return null;
    throw error;
  }
}

function currentPath(): string {
  return `${window.location.pathname}${window.location.search}`;
}

export function SessionProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const router = useRouter();
  const pathname = usePathname();

  const query = useQuery({
    queryKey: queryKeys.session.me,
    queryFn: fetchSession,
    // O provedor controla quando a sessão é rebuscada; uma falha de rede não
    // vira "sem sessão".
    staleTime: Infinity,
  });

  // Trava de redirecionamento: vários 401 (ou o portão e um 401) levam a um
  // único redirecionamento. É liberada quando o caminho muda.
  const redirecting = useRef(false);
  const loggingOut = useRef(false);
  useEffect(() => {
    redirecting.current = false;
  }, [pathname]);

  const navigateOnce = useCallback(
    (href: string) => {
      if (redirecting.current) return false;
      redirecting.current = true;
      router.replace(href);
      return true;
    },
    [router],
  );

  /**
   * Esquece a sessão: marca "sem sessão" e limpa todo o resto do cache
   * (consultas e mutações). A marca é gravada antes, para que quem observa a
   * sessão (o portão) desmonte as telas e nada seja buscado de novo.
   */
  const forgetSession = useCallback(() => {
    queryClient.setQueryData(queryKeys.session.me, null);
    queryClient.removeQueries({ predicate: (q) => q.queryHash !== SESSION_HASH });
    queryClient.getMutationCache().clear();
  }, [queryClient]);

  const hasSession = useCallback(
    () => queryClient.getQueryData<AuthContext | null>(queryKeys.session.me) != null,
    [queryClient],
  );

  useEffect(
    () =>
      configureApiClient({
        getCsrfToken: () =>
          queryClient.getQueryData<AuthContext | null>(queryKeys.session.me)?.csrf_token,
        onUnauthenticated: () => {
          // Sem sessão conhecida (primeira consulta, /entrar) não há sessão a
          // encerrar: o portão das rotas autenticadas decide.
          if (loggingOut.current || !hasSession()) return;
          if (navigateOnce(loginPath({ next: currentPath(), sessionEnded: true }))) {
            forgetSession();
          }
        },
        onPasswordChangeRequired: () => {
          if (window.location.pathname !== routes.contaSenha) navigateOnce(routes.contaSenha);
        },
        onCsrfInvalid: () => {
          // A API recusou a escrita antes de executá-la; ela não é repetida.
          void queryClient.invalidateQueries({ queryKey: queryKeys.session.me });
          toast.error(MESSAGE_CSRF_REFRESHED);
        },
      }),
    [queryClient, navigateOnce, forgetSession, hasSession],
  );

  const login = useCallback(
    async (email: string, password: string) => {
      let result;
      try {
        result = await api.identity.POST("/api/v1/auth/login", { body: { email, password } });
      } catch (cause) {
        throw new LoginError(networkError(cause), null);
      }
      if (!result.response.ok || !result.data) {
        throw new LoginError(
          toApiError(result.response, result.error),
          result.response.headers.get("Retry-After"),
        );
      }
      queryClient.setQueryData(queryKeys.session.me, result.data);
      return result.data;
    },
    [queryClient],
  );

  const logout = useCallback(async () => {
    loggingOut.current = true;
    try {
      const { response } = await api.identity.POST("/api/v1/auth/logout");
      // 401: a sessão já não existe no servidor; sair continua valendo.
      if (!response.ok && response.status !== 401) {
        toast.error(MESSAGE_LOGOUT_FAILED);
        return;
      }
    } catch {
      toast.error(MESSAGE_LOGOUT_FAILED);
      return;
    } finally {
      loggingOut.current = false;
    }
    redirecting.current = false;
    navigateOnce(routes.entrar);
    forgetSession();
  }, [navigateOnce, forgetSession]);

  const redirectToLogin = useCallback(() => {
    navigateOnce(loginPath({ next: currentPath() }));
  }, [navigateOnce]);

  const context = query.data ?? null;
  const status: SessionStatus =
    query.data === undefined ? "loading" : query.data === null ? "anonymous" : "authenticated";
  const error = query.data === undefined && isApiError(query.error) ? query.error : null;
  const { refetch: refetchQuery } = query;
  const refetch = useCallback(() => void refetchQuery(), [refetchQuery]);

  const value = useMemo<SessionValue>(
    () => ({ status, context, error, refetch, login, logout, redirectToLogin }),
    [status, context, error, refetch, login, logout, redirectToLogin],
  );

  return <SessionContext.Provider value={value}>{children}</SessionContext.Provider>;
}

export function useSession(): SessionValue {
  const value = useContext(SessionContext);
  if (!value) throw new Error("useSession precisa estar dentro de SessionProvider.");
  return value;
}

/** A pessoa tem a permissão efetiva? Nunca decide pelo nome do papel. */
export function usePermission(permission: string): boolean {
  const { context } = useSession();
  return context?.permissions.includes(permission) ?? false;
}

/** A pessoa tem todas as permissões efetivas pedidas? */
export function useCan(...permissions: string[]): boolean {
  const { context } = useSession();
  if (!context) return false;
  return permissions.every((p) => context.permissions.includes(p));
}
