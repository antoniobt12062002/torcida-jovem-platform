// Rotas sem sessão (login e recuperação de acesso): conteúdo centralizado.
export default function PublicLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="flex flex-1 items-center justify-center px-4 py-12">
      <div className="w-full max-w-sm">
        <p className="mb-8 text-center text-sm font-medium tracking-wide text-muted-foreground">
          TJ Platform
        </p>
        {children}
      </div>
    </main>
  );
}
