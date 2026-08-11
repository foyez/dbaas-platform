<script setup lang="ts">
import { useAuth } from 'vue-oidc-context'

const auth = useAuth()
</script>

<template>
  <button v-if="!auth.isAuthenticated" @click="auth.signinRedirect()">Log in</button>
  <button v-else @click="auth.signoutRedirect()">Log out</button>
</template>
