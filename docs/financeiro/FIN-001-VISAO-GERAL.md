# TJ PLATFORM
# FIN-001 — Visão Geral do Módulo Financeiro

Versão: 1.0

Status: Aprovado para desenvolvimento

---

# 1. Objetivo

O módulo financeiro da TJ Platform será responsável pela gestão financeira institucional da Torcida Jovem do Campo Mourão Futsal.

O módulo não deve ser tratado apenas como um controle de entradas e saídas.

Ele deve funcionar como um ERP financeiro da instituição, integrando:

- produtos;
- estoque;
- eventos;
- associados;
- fornecedores;
- prestação de contas;
- transparência pública.

---

# 2. Princípio Fundamental

O sistema deve representar operações reais da instituição.

Uma movimentação financeira nunca deve existir apenas como uma movimentação bancária.

Exemplo incorreto:

"PIX recebido João - R$60"

Exemplo correto:

"Venda de Copo Oficial TJ para João Silva - R$60"

---

# 3. Objetivos do módulo

O financeiro deve permitir:

- organização interna da diretoria;
- acompanhamento de resultados;
- controle financeiro institucional;
- transparência pública;
- prestação de contas;
- tomada de decisão.

---

# 4. Estrutura financeira

O sistema será baseado em:

## Receitas

Toda receita deve possuir uma origem identificada.

Principais categorias:

## Produtos

Exemplos:

- Camisetas;
- Copos;
- Bonés;
- Bandeiras;
- Outros produtos.

---

## Associados

Exemplos:

- Plano anual;
- Arquibancada;
- Outros planos.

---

## Eventos

Exemplos:

- Feijoada;
- Churrasco;
- Festas;
- Caravanas.

---

## Patrocínios

Registro de recursos recebidos de parceiros.

---

## Doações

Permite:

- doações identificadas;
- doações anônimas.

---

# 5. Despesas

Principais categorias:

## Produtos

- Compra fornecedor;
- Produção;
- Frete.

---

## Eventos

- Alimentação;
- Bebidas;
- Estrutura.

---

## Material torcida

- Bandeiras;
- Instrumentos;
- Materiais de arquibancada.

---

## Viagens

- Transporte;
- Hospedagem;
- Alimentação.

---

## Administrativo

- Sistemas;
- Contador;
- Taxas;
- Outros.

---

# 6. Plano de Contas

O plano de contas possui estrutura hierárquica.

Exemplo:

Receitas

    Produtos
        Camisetas
        Copos
        Bonés

    Associados
        Plano anual
        Arquibancada

    Eventos
        Feijoada
        Churrasco

    Patrocínios

    Doações


Despesas

    Produtos

    Eventos

    Material torcida

    Viagens

    Administrativo


Novas categorias podem ser criadas por diretores autorizados.

Categorias oficiais não devem ser excluídas definitivamente.

Devem ser desativadas.

---

# 7. Lançamento Financeiro

O lançamento financeiro é a principal entidade do módulo.

Cada lançamento deve possuir:

## Identificação

- ID;
- data criação;
- usuário criador;
- responsável pela operação.

---

## Classificação

- Tipo:
    - Receita;
    - Despesa.

- Categoria;
- Subcategoria;
- Centro de resultado.

---

## Valores

- Valor bruto;
- Taxas;
- Valor líquido.

---

## Pagamento

Formas aceitas:

- PIX;
- Cartão;
- Dinheiro;
- Transferência;
- Outros.

---

## Parcelamento

O sistema deve permitir parcelamentos.

Uma operação gera múltiplas parcelas.

Exemplo:

Venda camiseta:

Valor:
R$240

Pagamento:
3 parcelas

Sistema controla:

Parcela 1:
R$80

Parcela 2:
R$80

Parcela 3:
R$80

---

# 8. Taxas financeiras

Taxas devem ser controladas.

Exemplo:

Venda:
R$100

Taxa Mercado Pago:
R$4

Valor líquido:
R$96

O sistema deve diferenciar:

- valor da venda;
- custo financeiro;
- valor recebido.

---

# 9. Fornecedores

Fornecedor é uma entidade própria.

Cadastro:

- Nome;
- CNPJ;
- Contato;
- Produtos fornecidos;
- Histórico de compras;
- Valores praticados.

O sistema deve permitir consultar histórico.

Exemplo:

Fornecedor X:

2026:
100 camisetas

2027:
200 camisetas

---

# 10. Processo de compra

Compras devem seguir o fluxo:

Pedido de compra

↓

Fornecedor entrega

↓

Entrada no estoque

↓

Registro financeiro da despesa


---

# 11. Centro de Resultado

Toda operação pode estar vinculada a um centro de resultado.

Exemplos:

- Produto;
- Evento;
- Campanha;
- Ação específica.

O sistema deve permitir rateio.

Exemplo:

Compra TNT:

50% Final

50% Arquibancada.

---

# 12. Produtos e Financeiro

Venda de produto deve gerar automaticamente:

- baixa de estoque;
- receita financeira;
- histórico do comprador;
- relatório de vendas.

---

Compra de estoque deve gerar:

- entrada no estoque;
- despesa financeira.

---

# 13. Eventos

Eventos são entidades próprias.

Um evento possui:

- Nome;
- Data;
- Responsável;
- Orçamento;
- Receitas;
- Despesas;
- Participantes;
- Resultado financeiro.

Toda venda relacionada ao evento deve alimentar seu resultado.

Exemplo:

Feijoada TJ:

Receitas:

Ingressos:
R$5.000

Bebidas:
R$2.000


Despesas:

Alimentos:
R$3.000


Resultado:

R$4.000

---

# 14. Status dos lançamentos

## Receita

Estados:

Criada

↓

Confirmada

↓

Recebida

↓

Contabilizada


---

## Despesa

Estados:

Criada

↓

Pago

↓

Com comprovante

↓

Disponível para prestação

---

# 15. Documentos

O sistema deve possuir gestão documental.

Documentos aceitos:

- Nota fiscal;
- Recibo;
- Comprovante PIX;
- Contrato;
- Orçamento;
- Fotos.

Documentos ficam disponíveis apenas para:

- administração;
- conselho fiscal.

---

# 16. Prestação de contas

A prestação de contas será anual.

Fluxo:

Ano encerrado

↓

Sistema gera relatório

↓

Conselho fiscal analisa

↓

Parecer

↓

Aprovação

↓

Ano bloqueado


---

Após aprovação:

Alterações não podem modificar o passado.

Correções devem ocorrer através de:

Ajuste extraordinário.

O ajuste deve possuir:

- motivo;
- responsável;
- data;
- histórico.

---

# 17. Auditoria

O sistema deve registrar:

- criação;
- alteração;
- usuário responsável;
- data/hora;
- valor anterior;
- valor atualizado.

---

# 18. Exclusão

Não existe exclusão definitiva.

Lançamentos devem ser cancelados.

Exemplo:

Despesa cancelada:

Motivo:
Lançamento duplicado.

Responsável:
Maria.

---

# 19. Saldo financeiro

O sistema controla:

Saldo calculado:

Baseado nos lançamentos.

Também permite:

Conferência com saldo real.

Exemplo:

Saldo sistema:

R$5.000

Saldo informado banco:

R$4.700

Diferença:

R$300

---

# 20. Orçamento

O sistema deve permitir planejamento.

Exemplo:

Temporada 2027:

Receita prevista:

Produtos:
R$20.000

Eventos:
R$20.000

Associados:
R$10.000


Depois comparar:

Planejado x Realizado.

---

# 21. Portal Transparência

Área pública.

Exibição:

Por categoria.

Não exibir:

- dados pessoais;
- documentos;
- informações sensíveis.

Exibir:

Receitas:

- Produtos;
- Associados;
- Eventos;
- Patrocínios.


Despesas:

- Produtos;
- Eventos;
- Operacional.


Atualização:

Tempo real após validação.

---

# 22. Relatórios Internos

A diretoria deve possuir:

- Balancete mensal;
- Fluxo de caixa;
- Receita por categoria;
- Despesa por categoria;
- Resultado por produto;
- Resultado por evento;
- Histórico financeiro;
- Prestação de contas anual.

---

# 23. Exportações

Área administrativa:

Permitir exportação:

- PDF;
- Excel;
- CSV.

---

# 24. Permissões

## Presidente

Acesso completo.

---

## Diretor financeiro

Gestão financeira completa.

---

## Diretoria

Visualização e acompanhamento.

---

## Conselho fiscal

Consulta:

- documentos;
- lançamentos;
- prestação de contas.

Pode:

- emitir parecer;
- aprovar prestação.

---

## Associado

Acesso somente ao portal público.

---

# 25. Alertas

O sistema deve gerar alertas:

- despesas sem comprovante;
- prestação pendente;
- produto sem margem;
- fornecedor inexistente;
- inconsistências financeiras.

---

# 26. Responsividade

O módulo deve funcionar em dispositivos móveis.

Motivo:

A operação da torcida ocorre frequentemente em:

- jogos;
- eventos;
- vendas presenciais.

---

# Fim da especificação