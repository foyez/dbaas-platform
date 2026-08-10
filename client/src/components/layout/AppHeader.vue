<script setup lang="ts">
import { Button } from '@/components/ui/button'
</script>

<template>
  <header class="flex h-16 items-center justify-between border-b bg-background px-4 sm:px-6">
    <div class="md:hidden">
      <span class="font-semibold"> PostgreSQL PaaS </span>
    </div>

    <div class="hidden md:block" />

    <div class="flex items-center gap-2">
      <Button variant="ghost" size="sm"> Account </Button>
    </div>
  </header>
</template>
