<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { Database, Home, BookOpen, ScrollText } from 'lucide-vue-next'

import { RouterLink } from 'vue-router'
import { userManager } from '@/auth/oidc'
import { hasRole } from '@/auth/roles'

const isAdmin = ref(false)

onMounted(async () => {
  const user = await userManager.getUser()
  isAdmin.value = hasRole(user, 'dbaas.admin')
})
</script>

<template>
  <aside class="hidden w-64 shrink-0 border-r bg-background md:block">
    <div class="flex h-16 items-center border-b px-6">
      <RouterLink :to="{ name: 'dashboard' }" class="flex items-center gap-2 font-semibold">
        <Database class="h-5 w-5" />

        PostgreSQL PaaS
      </RouterLink>
    </div>

    <nav class="space-y-1 p-4">
      <RouterLink
        :to="{ name: 'dashboard' }"
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
        :to="{ name: 'instances' }"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        active-class="
          bg-muted
          text-foreground
        "
      >
        <Database class="h-4 w-4" />

        Instances
      </RouterLink>

      <RouterLink
        v-if="isAdmin"
        :to="{ name: 'audit-logs' }"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
        active-class="
          bg-muted
          text-foreground
        "
      >
        <ScrollText class="h-4 w-4" />

        Audit log
      </RouterLink>
    </nav>

    <div class="mt-4 border-t p-4">
      <RouterLink
        :to="{ name: 'docs-overview' }"
        class="flex items-center gap-3 rounded-md px-3 py-2 text-sm text-muted-foreground hover:bg-muted hover:text-foreground"
      >
        <BookOpen class="h-4 w-4" />

        Documentation
      </RouterLink>
    </div>
  </aside>
</template>