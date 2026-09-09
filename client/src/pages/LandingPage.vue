<script setup lang="ts">
import { onMounted } from 'vue'
import { useAuth } from 'vue-oidc-context'
import { useRouter } from 'vue-router'
import { Database, Terminal, ShieldCheck, Activity, GitBranch, KeyRound } from 'lucide-vue-next'

import { Button } from '@/components/ui/button'

const auth = useAuth()
const router = useRouter()

// A returning, already-authenticated visitor shouldn't land on marketing
// copy every time - send them straight to their instances instead.
onMounted(() => {
  if (auth.value.isAuthenticated) {
    router.replace({ name: 'dashboard' })
  }
})

function getStarted() {
  if (auth.value.isAuthenticated) {
    router.push({ name: 'dashboard' })
  } else {
    auth.value.signinRedirect()
  }
}

const features = [
  {
    icon: Database,
    title: 'Managed PostgreSQL',
    description:
      'Create a production PostgreSQL instance in a few clicks - provisioning, health checks, and lifecycle management are handled for you.',
  },
  {
    icon: Terminal,
    title: 'Built-in query console',
    description: 'Run SQL directly in the browser - no local client or connection setup required.',
  },
  {
    icon: KeyRound,
    title: 'Ready-to-use credentials',
    description:
      'Each instance gets a real, externally-reachable connection endpoint - copy the connection string and go.',
  },
  {
    icon: ShieldCheck,
    title: 'OIDC authentication',
    description: 'Sign in once; every instance is isolated to your account and enforced on every request.',
  },
  {
    icon: Activity,
    title: 'Logs & audit trail',
    description: 'Per-instance logs and an admin audit log, backed by Prometheus, Loki, and Grafana.',
  },
  {
    icon: GitBranch,
    title: 'A real Go SDK',
    description: 'Everything in the UI is also available over a REST API and a typed Go SDK.',
  },
]
</script>

<template>
  <main class="min-h-screen bg-background">
    <!-- Navigation -->
    <header class="border-b">
      <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <RouterLink :to="{ name: 'landing' }" class="flex items-center gap-2 font-semibold">
          <Database class="h-5 w-5" />
          PostgreSQL PaaS
        </RouterLink>

        <div class="flex items-center gap-2">
          <RouterLink
            :to="{ name: 'docs-overview' }"
            class="px-3 py-2 text-sm text-muted-foreground hover:text-foreground"
          >
            Documentation
          </RouterLink>

          <Button v-if="!auth.isAuthenticated" variant="outline" @click="auth.signinRedirect()">
            Log in
          </Button>

          <Button v-else variant="outline" @click="auth.signoutRedirect()"> Log out </Button>
        </div>
      </div>
    </header>

    <!-- Hero -->
    <section class="mx-auto flex max-w-6xl flex-col items-center px-6 py-24 text-center">
      <div class="max-w-3xl">
        <p class="mb-4 text-sm font-medium text-muted-foreground">Database as a Service</p>

        <h1 class="text-4xl font-bold tracking-tight sm:text-6xl">
          Managed PostgreSQL,
          <span class="text-muted-foreground">without the infrastructure work.</span>
        </h1>

        <p class="mx-auto mt-6 max-w-2xl text-lg leading-8 text-muted-foreground">
          Provision a production-ready PostgreSQL instance in minutes. We handle Kubernetes,
          networking, and lifecycle management so you don't have to.
        </p>

        <div class="mt-8 flex justify-center gap-3">
          <Button size="lg" @click="getStarted">
            {{ auth.isAuthenticated ? 'Go to dashboard' : 'Get started' }}
          </Button>

          <Button variant="outline" size="lg" @click="router.push({ name: 'docs-overview' })">
            Documentation
          </Button>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section class="border-t bg-muted/30">
      <div class="mx-auto grid max-w-6xl gap-8 px-6 py-16 sm:grid-cols-2 lg:grid-cols-3">
        <div v-for="feature in features" :key="feature.title">
          <component :is="feature.icon" class="mb-3 h-5 w-5 text-muted-foreground" />
          <h2 class="font-semibold">{{ feature.title }}</h2>
          <p class="mt-2 text-sm text-muted-foreground">{{ feature.description }}</p>
        </div>
      </div>
    </section>
  </main>
</template>