<script setup lang="ts">
import { Database, Home, BookOpen } from 'lucide-vue-next'

import { RouterLink } from 'vue-router'
</script>

<template>
  <aside class="hidden w-64 shrink-0 border-r bg-background md:block">
    <div class="flex h-16 items-center border-b px-6">
      <RouterLink to="/" class="flex items-center gap-2 font-semibold">
        <Database class="h-5 w-5" />

        PostgreSQL PaaS
      </RouterLink>
    </div>

    <nav class="space-y-1 p-4">
      <RouterLink
        to="/"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        active-class="
          bg-muted
          text-foreground
        "
      >
        <Home class="h-4 w-4" />

        Dashboard
      </RouterLink>

      <RouterLink
        to="/instances"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        active-class="
          bg-muted
          text-foreground
        "
      >
        <Database class="h-4 w-4" />

        Instances
      </RouterLink>
    </nav>

    <div class="mt-4 border-t p-4">
      <a
        href="#"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        <BookOpen class="h-4 w-4" />

        Documentation
      </a>
    </div>
  </aside>
</template>
