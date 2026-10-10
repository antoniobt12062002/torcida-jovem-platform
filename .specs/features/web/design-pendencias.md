# Pendências de design da Web V1

Levantadas na conferência visual de 2026-10-09 (`smoke-visual.md`). São questões de aparência e de UI, consequência de as specs da V1 cobrirem comportamento, permissões, erros e acessibilidade, e não identidade visual. **Nenhuma foi corrigida na Web V1**: ficam como entrada do ciclo próprio de identidade visual e UI/UX.

| # | Pendência | Onde |
| --- | --- | --- |
| D-01 | A fonte de toda a interface cai em Times New Roman. `web/app/globals.css` define `--font-sans: var(--font-sans)`, uma referência a si mesma, em vez de apontar para a fonte carregada no layout (`--font-geist-sans`). Já existe na `develop`; a Web V1 não alterou o arquivo. | Todas as telas |
| D-02 | Sem identidade visual: paleta só de preto, branco e cinza, sem as cores da Torcida Jovem, sem escudo nem logo. A marca é o texto "TJ Platform". | Shell, login, páginas públicas |
| D-03 | Hierarquia fraca: títulos pequenos, pouco contraste entre seções, e larguras de página diferentes entre áreas (início, contas, saldo, lançamentos). | Todas as telas |
| D-04 | Tabelas no celular só rolam na horizontal; não há layout em cartões. Colunas de texto longo, como o motivo, quebram em muitas linhas. | Lançamentos, movimentações, usuários, comprovantes |
| D-05 | Ações repetidas com o mesmo nome acessível em todas as linhas: "Devolver" (movimentações) e "Detalhes" (lançamentos). O nome deveria identificar a linha, como já acontece em contas e usuários. | Movimentações, lançamentos |
| D-06 | Estados vazios só com uma frase, sem destaque nem chamada para a ação. | Contas, lançamentos, produtos, comprovantes |
| D-07 | Campos de seleção nativos com aparência diferente dos demais campos. | Filtros de usuários e lançamentos, formulários de lançamento |
| D-08 | Lançamentos identificados por código curto no título ("Receita fb6e3d8f"). Ligado a WEB-D-014. | Detalhe do lançamento, referência de devolução |
| D-09 | O limite de anexo aparece como "até 9,9 MB", número pouco natural para quem usa. | Comprovantes |
| D-10 | Tela de erro de renderização em inglês e sem a identidade do sistema. | Global (`app/error.tsx` não existe) |
| D-11 | Sem modo escuro selecionável (fora do V1 por decisão registrada). A decidir no ciclo de design. | Global |
