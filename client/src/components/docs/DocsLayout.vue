<script setup lang="ts">
import { Database } from 'lucide-vue-next'

const sections = [
  { to: { name: 'docs-overview' }, label: 'Overview' },
  { to: { name: 'docs-getting-started' }, label: 'Getting started' },
  { to: { name: 'docs-architecture' }, label: 'Architecture' },
  { to: { name: 'docs-api' }, label: 'API reference' },
  { to: { name: 'docs-sdk' }, label: 'Go SDK' },
]
</script>

<template>
  <div class="min-h-screen bg-background">
    <header class="border-b">
      <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <RouterLink :to="{ name: 'landing' }" class="flex items-center gap-2 font-semibold">
          <Database class="h-5 w-5" />
          PostgreSQL PaaS
        </RouterLink>

        <RouterLink :to="{ name: 'dashboard' }" class="text-sm text-muted-foreground hover:text-foreground">
          Go to dashboard
        </RouterLink>
      </div>
    </header>

    <div class="mx-auto flex max-w-6xl gap-10 px-6 py-10">
      <aside class="hidden w-56 shrink-0 md:block">
        <nav class="sticky top-10 space-y-1">
          <RouterLink
            v-for="section in sections"
            :key="section.label"
            :to="section.to"
            class="block rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
            active-class="bg-muted text-foreground"
          >
            {{ section.label }}
          </RouterLink>
        </nav>
      </aside>

      <main class="min-w-0 max-w-2xl flex-1 pb-24">
        <RouterView />
      </main>
    </div>
  </div>
</template>