<script setup lang="ts">
import InstanceFormFields from './InstanceFormFields.vue'

import type { UpdateInstanceInput } from '@/schemas/instance'

withDefaults(
  defineProps<{
    initialValue?: Partial<UpdateInstanceInput>
    loading?: boolean
  }>(),
  {
    initialValue: undefined,
    loading: false,
  },
)

const emit = defineEmits<{
  submit: [value: UpdateInstanceInput]
}>()
</script>

<template>
  <InstanceFormFields
    mode="edit"
    :initial-value="initialValue"
    :loading="loading"
    @submit="emit('submit', $event as UpdateInstanceInput)"
  />
</template>
