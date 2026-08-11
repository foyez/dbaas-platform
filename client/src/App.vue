<script setup lang="ts">
import AuthProvider from 'vue-oidc-context'
import { oidcSettings, onSigninCallback } from './auth/oidc'
</script>

<template>
  <AuthProvider v-bind="oidcSettings">
    <RouterView />
  </AuthProvider>
</template>

<style scoped></style>
