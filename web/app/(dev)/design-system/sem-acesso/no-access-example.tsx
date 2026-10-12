"use client";

import { PageContainer } from "@/components/app/page-container";
import { NoAccess } from "@/components/app/require-permission";

import { ExampleShell } from "../example-shell";

export function NoAccessExample() {
  return (
    <ExampleShell>
      <PageContainer width="narrow">
        <NoAccess />
      </PageContainer>
    </ExampleShell>
  );
}
