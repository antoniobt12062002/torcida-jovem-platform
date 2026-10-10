import { redirect } from "next/navigation";

import { routes } from "@/lib/session/routes";

// A raiz leva à página inicial autenticada (INT-01 AC3). Sem sessão, o portão
// de app/(app)/layout.tsx leva a /entrar.
export default function Home(): never {
  redirect(routes.inicio);
}
