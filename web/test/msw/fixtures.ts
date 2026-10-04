import { HttpResponse } from "msw";

import type { components as Estoque } from "@/lib/api/estoque";
import type { components as Financeiro } from "@/lib/api/financeiro";
import type { components as Identity } from "@/lib/api/identity";

// Fábricas de dados de teste, tipadas pelos contratos gerados. Os valores são
// fictícios (nenhum dado pessoal real). Cada fábrica aceita sobrescritas.

export type AuthContext = Identity["schemas"]["AuthContext"];
export type User = Identity["schemas"]["User"];
export type Problem = Identity["schemas"]["Problem"];
export type FieldError = Identity["schemas"]["FieldError"];
export type Conta = Financeiro["schemas"]["Conta"];
export type Lancamento = Financeiro["schemas"]["Lancamento"];
export type Comprovante = Financeiro["schemas"]["Comprovante"];
export type SaldoFinanceiro = Financeiro["schemas"]["Saldo"];
export type Produto = Estoque["schemas"]["Produto"];
export type Movimentacao = Estoque["schemas"]["Movimentacao"];
export type SaldoEstoque = Estoque["schemas"]["Saldo"];

export const ids = {
  user: "00000000-0000-4000-8000-000000000001",
  otherUser: "00000000-0000-4000-8000-000000000002",
  conta: "00000000-0000-4000-8000-000000000101",
  lancamento: "00000000-0000-4000-8000-000000000201",
  comprovante: "00000000-0000-4000-8000-000000000301",
  produto: "00000000-0000-4000-8000-000000000401",
  movimentacao: "00000000-0000-4000-8000-000000000501",
} as const;

const NOW = "2026-10-04T12:00:00Z";

export function authContext(overrides: Partial<AuthContext> = {}): AuthContext {
  return {
    user: { id: ids.user, email: "pessoa.teste@example.com", name: "Pessoa Teste" },
    roles: [],
    permissions: [],
    csrf_token: "csrf-token-de-teste",
    must_change_password: false,
    admin_membership: null,
    ...overrides,
  };
}

export function user(overrides: Partial<User> = {}): User {
  return {
    id: ids.otherUser,
    email: "outra.pessoa@example.com",
    name: "Outra Pessoa",
    active: true,
    must_change_password: false,
    created_at: NOW,
    roles: ["ASSOCIADO"],
    admin_membership: null,
    ...overrides,
  };
}

export function problemBody(
  status: number,
  code: string,
  overrides: Partial<Problem> = {},
): Problem {
  return {
    type: "about:blank",
    title: `HTTP ${status}`,
    status,
    detail: `detalhe interno de ${code}`,
    code,
    request_id: "req-teste",
    ...overrides,
  };
}

/** Resposta `application/problem+json`, como a API devolve. */
export function problem(
  status: number,
  code: string,
  overrides: Partial<Problem> = {},
  headers: Record<string, string> = {},
) {
  return HttpResponse.json(problemBody(status, code, overrides), {
    status,
    headers: { "Content-Type": "application/problem+json", ...headers },
  });
}

export function conta(overrides: Partial<Conta> = {}): Conta {
  return {
    id: ids.conta,
    tipo: "RECEITA",
    nome: "Mensalidades",
    parent_id: null,
    ativo: true,
    created_at: NOW,
    ...overrides,
  };
}

export function lancamento(overrides: Partial<Lancamento> = {}): Lancamento {
  return {
    id: ids.lancamento,
    tipo: "RECEITA",
    conta_id: ids.conta,
    valor_bruto_cents: 10000,
    taxa_cents: 0,
    valor_liquido_cents: 10000,
    forma_pagamento: "PIX",
    status: "CRIADA",
    devolucao_de_id: null,
    motivo_cancelamento: null,
    cancelado_por: null,
    cancelado_em: null,
    criado_por: ids.user,
    criado_em: NOW,
    atualizado_em: NOW,
    ...overrides,
  };
}

export function comprovante(overrides: Partial<Comprovante> = {}): Comprovante {
  return {
    id: ids.comprovante,
    original_filename: "comprovante.pdf",
    content_type: "application/pdf",
    size_bytes: 1024,
    version: 1,
    uploaded_by: ids.user,
    uploaded_at: NOW,
    ...overrides,
  };
}

export function saldoFinanceiro(saldo_cents = 0): SaldoFinanceiro {
  return { saldo_cents };
}

export function produto(overrides: Partial<Produto> = {}): Produto {
  return {
    id: ids.produto,
    codigo: "CAM-001",
    nome: "Camisa oficial",
    unidade_medida: "UN",
    criado_em: NOW,
    ...overrides,
  };
}

export function movimentacao(overrides: Partial<Movimentacao> = {}): Movimentacao {
  return {
    id: ids.movimentacao,
    produto_id: ids.produto,
    tipo: "ENTRADA",
    quantidade: 10,
    origem: "INVENTARIO",
    motivo: null,
    movimentacao_de_id: null,
    responsavel_id: ids.user,
    criado_em: NOW,
    ...overrides,
  };
}

export function saldoEstoque(saldo = 0): SaldoEstoque {
  return { saldo };
}
