<script setup lang="ts">
import { useAuth } from 'vue-oidc-context'
import { Button } from '@/components/ui/button'

const auth = useAuth()
</script>

<template>
  <header class="flex h-16 items-center justify-between border-b bg-background px-4 sm:px-6">
    <div class="md:hidden">
      <span class="font-semibold"> PostgreSQL PaaS </span>
    </div>
    <div class="hidden md:block" />
    <div class="flex items-center gap-2">
      <span v-if="auth.user?.profile.name" class="hidden text-sm text-muted-foreground sm:inline">
        {{ auth.user.profile.name }}
      </span>
      <Button v-if="auth.isAuthenticated" variant="ghost" size="sm" @click="auth.signoutRedirect()">
        Log out
      </Button>
    </div>
  </header>
</template>
