import type { components } from "@/lib/api/financeiro";

// Plano de contas em árvore por parent_id (FWB-01 AC1). Função pura: o dado
// guardado no cache continua a ContaList da API; a árvore é derivada na tela.

type Conta = components["schemas"]["Conta"];

export type ContaNode = {
  conta: Conta;
  children: ContaNode[];
};

/**
 * Monta a árvore mantendo a ordem da lista. Contas inativas entram como as
 * demais; uma conta cujo pai não está na lista vai para a raiz.
 */
export function buildContaTree(contas: readonly Conta[]): ContaNode[] {
  const nodes = new Map<string, ContaNode>();
  for (const conta of contas) nodes.set(conta.id, { conta, children: [] });

  const roots: ContaNode[] = [];
  for (const conta of contas) {
    const node = nodes.get(conta.id)!;
    const parent = conta.parent_id ? nodes.get(conta.parent_id) : undefined;
    if (parent) parent.children.push(node);
    else roots.push(node);
  }
  return roots;
}
