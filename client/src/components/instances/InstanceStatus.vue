<script setup lang="ts">
import { computed } from 'vue'

import { Badge } from '@/components/ui/badge'

import type { InstanceStatus } from '@/types/instance'

const props = defineProps<{
  status: InstanceStatus
}>()

const variant = computed(() => {
  switch (props.status) {
    case 'Running':
      return 'default'

    case 'Failed':
      return 'destructive'

    case 'Pending':
    case 'Provisioning':
    case 'Deleting':
      return 'secondary'

    default:
      return 'outline'
  }
})
</script>

<template>
  <Badge :variant="variant">
    {{ status }}
  </Badge>
</template>
