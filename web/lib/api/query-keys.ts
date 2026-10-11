import type { paths as IdentityPaths } from "./identity";

// Chaves do TanStack Query para todos os recursos do contrato (WEB-D-010).
// Frentes paralelas leem e invalidam o mesmo cache por aqui, sem importar
// código umas das outras. Cada chave começa pelo prefixo do módulo, para
// invalidar um módulo inteiro com `all`.

type ListUsersQuery = NonNullable<IdentityPaths["/api/v1/users"]["get"]["parameters"]["query"]>;

/** Filtros da lista de usuários (o cursor da página fica fora da chave). */
export type UsersFilters = Omit<ListUsersQuery, "cursor">;

const identityAll = ["identity"] as const;
const usersAll = [...identityAll, "users"] as const;
const financeiroAll = ["financeiro"] as const;
const estoqueAll = ["estoque"] as const;

export const queryKeys = {
  session: {
    me: ["session", "me"] as const,
  },
  identity: {
    all: identityAll,
    usersAll,
    users: (filters: UsersFilters = {}) => [...usersAll, filters] as const,
  },
  financeiro: {
    all: financeiroAll,
    contas: [...financeiroAll, "contas"] as const,
    lancamentos: [...financeiroAll, "lancamentos"] as const,
    saldo: [...financeiroAll, "saldo"] as const,
    comprovantes: (lancamentoId: string) =>
      [...financeiroAll, "comprovantes", lancamentoId] as const,
  },
  estoque: {
    all: estoqueAll,
    produtos: [...estoqueAll, "produtos"] as const,
    movimentacoes: (produtoId: string) =>
      [...estoqueAll, "movimentacoes", produtoId] as const,
    saldo: (produtoId: string) => [...estoqueAll, "saldo", produtoId] as const,
  },
} as const;
