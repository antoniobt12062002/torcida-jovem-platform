import {
  ArrowDownToLineIcon,
  ArrowUpFromLineIcon,
  BanIcon,
  CircleDashedIcon,
  HouseIcon,
  ListTreeIcon,
  PackageIcon,
  ReceiptIcon,
  TriangleAlertIcon,
  UsersIcon,
  WalletIcon,
} from "lucide-react";

import type { NavEntry } from "@/components/app/nav";
import type { StatusDef } from "@/components/app/status-badge";

// Dados de exemplo do espécime: fictícios, sem nome real nem dado pessoal.

export type StatusLancamento = "CRIADA" | "RECEBIDA" | "PAGA" | "CANCELADA";

/** Mapa de status de lançamento (DD-3): `CANCELADA` é neutra, porque cancelamento não é erro. */
export const STATUS_LANCAMENTO: Record<StatusLancamento, StatusDef> = {
  CRIADA: { label: "Criada", tone: "neutral", icon: CircleDashedIcon },
  RECEBIDA: { label: "Recebida", tone: "success", icon: ArrowDownToLineIcon },
  PAGA: { label: "Paga", tone: "success", icon: ArrowUpFromLineIcon },
  CANCELADA: { label: "Cancelada", tone: "neutral", icon: BanIcon },
};

export const SALDO_NEGATIVO: StatusDef = { label: "Saldo negativo", tone: "error", icon: TriangleAlertIcon };

export type Tipo = "RECEITA" | "DESPESA";

export const TIPO_LABEL: Record<Tipo, string> = { RECEITA: "Receita", DESPESA: "Despesa" };

export type Lancamento = {
  id: string;
  descricao: string;
  tipo: Tipo;
  conta: string;
  data: string;
  status: StatusLancamento;
  valorCents: number;
};

export const LANCAMENTOS: readonly Lancamento[] = [
  { id: "fb6e3d8f", descricao: "Mensalidades de março", tipo: "RECEITA", conta: "Associações", data: "02/03/2026", status: "RECEBIDA", valorCents: 184000 },
  { id: "a17c02e4", descricao: "Compra de faixas e bandeirões para a temporada", tipo: "DESPESA", conta: "Materiais", data: "04/03/2026", status: "PAGA", valorCents: 92050 },
  { id: "3d9b51c0", descricao: "Venda de camisetas na partida de sábado", tipo: "RECEITA", conta: "Loja", data: "07/03/2026", status: "RECEBIDA", valorCents: 268000 },
  { id: "c4e8a213", descricao: "Ônibus para a partida fora de casa", tipo: "DESPESA", conta: "Viagens", data: "09/03/2026", status: "CRIADA", valorCents: 145000 },
  { id: "7f20d6b9", descricao: "Rifa beneficente", tipo: "RECEITA", conta: "Eventos", data: "11/03/2026", status: "CANCELADA", valorCents: 50000 },
  { id: "9a5e7c31", descricao: "Impressão de ingressos", tipo: "DESPESA", conta: "Eventos", data: "12/03/2026", status: "PAGA", valorCents: 18990 },
  { id: "e2b4f870", descricao: "Doação de associado", tipo: "RECEITA", conta: "Doações", data: "14/03/2026", status: "RECEBIDA", valorCents: 30000 },
  { id: "5c81a9d2", descricao: "Manutenção dos instrumentos da bateria", tipo: "DESPESA", conta: "Materiais", data: "16/03/2026", status: "CRIADA", valorCents: 37500 },
  { id: "b06d3e45", descricao: "Venda de bonés", tipo: "RECEITA", conta: "Loja", data: "18/03/2026", status: "RECEBIDA", valorCents: 72000 },
  { id: "1e9f4a7c", descricao: "Taxa bancária", tipo: "DESPESA", conta: "Banco", data: "20/03/2026", status: "PAGA", valorCents: 2490 },
  { id: "d7a2c5e8", descricao: "Festa de aniversário da torcida", tipo: "DESPESA", conta: "Eventos", data: "23/03/2026", status: "CANCELADA", valorCents: 120000 },
  { id: "4b3e8f16", descricao: "Mensalidades atrasadas de fevereiro", tipo: "RECEITA", conta: "Associações", data: "27/03/2026", status: "CRIADA", valorCents: 46000 },
];

export type Comprovante = { id: string; arquivo: string; enviadoEm: string };

export const COMPROVANTES: readonly Comprovante[] = [
  { id: "cp1", arquivo: "nota-faixas.pdf", enviadoEm: "04/03/2026" },
  { id: "cp2", arquivo: "recibo-onibus.jpg", enviadoEm: "09/03/2026" },
  { id: "cp3", arquivo: "boleto-ingressos.pdf", enviadoEm: "12/03/2026" },
];

/** Navegação de exemplo com os ícones do design. Lançamentos leva à página de exemplo; os demais voltam ao espécime. */
export const NAV_EXEMPLO: readonly NavEntry[] = [
  { label: "Início", href: "/design-system?item=inicio", icon: HouseIcon },
  {
    label: "Financeiro",
    items: [
      { label: "Lançamentos", href: "/design-system/pagina-de-exemplo", icon: ReceiptIcon },
      { label: "Contas", href: "/design-system?item=contas", icon: ListTreeIcon },
      { label: "Saldo", href: "/design-system?item=saldo", icon: WalletIcon },
    ],
  },
  { label: "Estoque", items: [{ label: "Produtos", href: "/design-system?item=produtos", icon: PackageIcon }] },
  { label: "Administração", items: [{ label: "Usuários", href: "/design-system?item=usuarios", icon: UsersIcon }] },
];

export const PESSOA_EXEMPLO = { nome: "Pessoa de Exemplo", email: "pessoa@exemplo.com" };

export const PAPEIS_EXEMPLO = ["Tesouraria", "Diretoria", "Estoque e loja", "Administração do sistema"] as const;
