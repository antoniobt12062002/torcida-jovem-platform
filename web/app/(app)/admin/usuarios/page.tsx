import type { Metadata } from "next";

import { RequirePermission } from "@/components/app/require-permission";
import { UsersPage } from "@/features/identity/usuarios/users-page";

export const metadata: Metadata = {
  title: "Usuários",
};

export default function UsuariosPage() {
  return (
    <main className="mx-auto flex w-full max-w-6xl flex-1 flex-col px-4 py-8">
      <RequirePermission permission="identity:user:read" mode="page">
        <UsersPage />
      </RequirePermission>
    </main>
  );
}
