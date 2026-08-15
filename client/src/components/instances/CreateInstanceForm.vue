<script setup lang="ts">
import InstanceFormFields from './InstanceFormFields.vue'

import type { CreateInstanceInput } from '@/schemas/instance'

withDefaults(
  defineProps<{
    initialValue?: Partial<CreateInstanceInput>
    loading?: boolean
  }>(),
  {
    initialValue: undefined,
    loading: false,
  },
)

const emit = defineEmits<{
  submit: [value: CreateInstanceInput]
}>()
</script>

<template>
  <InstanceFormFields
    mode="create"
    :initial-value="initialValue"
    :loading="loading"
    @submit="emit('submit', $event as CreateInstanceInput)"
  />
</template>
