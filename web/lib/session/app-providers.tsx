"use client";

import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { type ReactNode, useState } from "react";

import { Toaster } from "@/components/ui/toast";
import { isApiError } from "@/lib/api/problem";

import { SessionProvider } from "./session-provider";

const MAX_RETRIES = 2;

/** Repete só falhas transitórias: rede e 5xx. Erros 4xx são respostas finais. */
export function shouldRetry(failureCount: number, error: unknown): boolean {
  if (failureCount >= MAX_RETRIES) return false;
  if (isApiError(error) && error.status >= 400 && error.status < 500) return false;
  return true;
}

export function createQueryClient() {
  return new QueryClient({
    defaultOptions: {
      queries: { retry: shouldRetry },
      // Escritas nunca são repetidas automaticamente.
      mutations: { retry: false },
    },
  });
}

/** Providers da raiz: consultas, sessão e avisos (toaster). */
export function AppProviders({ children }: { children: ReactNode }) {
  const [queryClient] = useState(createQueryClient);
  return (
    <QueryClientProvider client={queryClient}>
      <SessionProvider>
        {children}
        <Toaster />
      </SessionProvider>
    </QueryClientProvider>
  );
}
