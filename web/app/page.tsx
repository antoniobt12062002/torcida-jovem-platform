import { Badge } from "@/components/ui/badge";

export const dynamic = "force-dynamic";

async function apiOnline(): Promise<boolean> {
  try {
    const res = await fetch(`${process.env.API_URL ?? "http://localhost:8080"}/healthz`, {
      cache: "no-store",
    });
    return res.ok;
  } catch {
    return false;
  }
}

export default async function Home() {
  const online = await apiOnline();
  return (
    <main className="flex min-h-screen flex-col items-center justify-center gap-4 p-6">
      <h1 className="text-3xl font-bold">TJ Platform</h1>
      <Badge variant={online ? "default" : "destructive"}>
        {online ? "API online" : "API indisponível"}
      </Badge>
    </main>
  );
}
