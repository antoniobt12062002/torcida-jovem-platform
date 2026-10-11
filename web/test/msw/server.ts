import { setupServer } from "msw/node";

// Servidor MSW compartilhado pelos testes. Sem handlers padrão: cada teste
// registra os seus com server.use(...), descartados ao fim do teste.
export const server = setupServer();
