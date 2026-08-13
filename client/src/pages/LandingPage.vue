<script setup lang="ts">
import { useAuth } from 'vue-oidc-context'
import { Button } from '@/components/ui/button'
import { Database } from 'lucide-vue-next'

const auth = useAuth()
</script>

<template>
  <main class="min-h-screen bg-background">
    <!-- Navigation -->
    <header class="border-b">
      <div class="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <div class="flex h-16 items-center border-b px-6">
          <RouterLink to="/" class="flex items-center gap-2 font-semibold">
            <Database class="h-5 w-5" />

            PostgreSQL PaaS
          </RouterLink>
        </div>

        <Button v-if="!auth.isAuthenticated" @click="auth.signinRedirect()"> Log in </Button>

        <Button v-else variant="outline" @click="auth.signoutRedirect()"> Log out </Button>
      </div>
    </header>

    <!-- Hero -->
    <section class="mx-auto flex max-w-6xl flex-col items-center px-6 py-24 text-center">
      <div class="max-w-3xl">
        <p class="mb-4 text-sm font-medium text-muted-foreground">Platform as a Service</p>

        <h1 class="text-4xl font-bold tracking-tight sm:text-6xl">
          Deploy your applications.
          <span class="text-muted-foreground">We handle the infrastructure.</span>
        </h1>

        <p class="mx-auto mt-6 max-w-2xl text-lg leading-8 text-muted-foreground">
          A simple platform for deploying and managing your applications without worrying about
          Kubernetes, networking, or infrastructure.
        </p>

        <div class="mt-8 flex justify-center gap-3">
          <Button v-if="!auth.isAuthenticated" size="lg" @click="auth.signinRedirect()">
            Get started
          </Button>

          <Button v-else size="lg" @click="$router.push('/dashboard')"> Go to dashboard </Button>

          <Button variant="outline" size="lg"> Documentation </Button>
        </div>
      </div>
    </section>

    <!-- Features -->
    <section class="border-t bg-muted/30">
      <div class="mx-auto grid max-w-6xl gap-8 px-6 py-16 sm:grid-cols-3">
        <div>
          <h2 class="font-semibold">Simple deployments</h2>
          <p class="mt-2 text-sm text-muted-foreground">
            Deploy your application without managing Kubernetes yourself.
          </p>
        </div>

        <div>
          <h2 class="font-semibold">Built for developers</h2>
          <p class="mt-2 text-sm text-muted-foreground">
            Focus on your code while the platform takes care of the infrastructure.
          </p>
        </div>

        <div>
          <h2 class="font-semibold">Production ready</h2>
          <p class="mt-2 text-sm text-muted-foreground">
            Reliable infrastructure with networking, scaling, and security built in.
          </p>
        </div>
      </div>
    </section>
  </main>
</template>
