<script setup lang="ts">
import { watch } from 'vue'
import { useAuth } from 'vue-oidc-context'
import { useRouter } from 'vue-router'

const auth = useAuth()
const router = useRouter()

watch(
  auth,
  () => {
    if (!auth.value.isLoading && auth.value.isAuthenticated) {
      router.replace({ name: 'dashboard' })
    }
  },
  { immediate: true },
)
</script>

<template>
  <p>Signing you in…</p>
</template>
