// Busca de produtos por código ou nome, no cliente (EWB-01 AC1, WEB-D-016).
// Sem diferença de maiúsculas nem de acentos.

type Searchable = { codigo: string; nome: string };

function normalize(text: string): string {
  return text.normalize("NFD").replace(/\p{Diacritic}/gu, "").toLowerCase().trim();
}

export function searchProdutos<T extends Searchable>(produtos: readonly T[], term: string): T[] {
  const needle = normalize(term);
  if (needle === "") return [...produtos];
  return produtos.filter(
    (p) => normalize(p.codigo).includes(needle) || normalize(p.nome).includes(needle),
  );
}
