<script setup lang="ts">
import { ref, watch } from 'vue'

import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'

import { createInstanceSchema, updateInstanceSchema } from '@/schemas/instance'

import type { CreateInstanceInput, UpdateInstanceInput } from '@/schemas/instance'

const props = withDefaults(
  defineProps<{
    mode: 'create' | 'edit'
    initialValue?: Partial<CreateInstanceInput>
    loading?: boolean
  }>(),
  {
    initialValue: undefined,
    loading: false,
  },
)

const emit = defineEmits<{
  submit: [value: CreateInstanceInput | UpdateInstanceInput]
}>()

const form = ref<CreateInstanceInput>({
  name: props.initialValue?.name ?? '',
  instances: props.initialValue?.instances ?? 1,
  version: props.initialValue?.version ?? 16,
  storage: props.initialValue?.storage ?? '10Gi',
})

const errors = ref<Partial<Record<keyof CreateInstanceInput, string>>>({})

watch(
  () => props.initialValue,
  (value) => {
    if (!value) {
      return
    }

    form.value = {
      name: value.name ?? '',
      instances: value.instances ?? 1,
      version: value.version ?? 16,
      storage: value.storage ?? '10Gi',
    }
  },
  { deep: true },
)

function submit() {
  const schema = props.mode === 'edit' ? updateInstanceSchema : createInstanceSchema

  const result = schema.safeParse(form.value)

  if (!result.success) {
    const fieldErrors: typeof errors.value = {}

    for (const issue of result.error.issues) {
      const field = issue.path[0]

      if (typeof field === 'string' && field in form.value) {
        fieldErrors[field as keyof CreateInstanceInput] = issue.message
      }
    }

    errors.value = fieldErrors
    return
  }

  errors.value = {}

  emit('submit', result.data)
}
</script>

<template>
  <form class="space-y-6" @submit.prevent="submit">
    <!-- Create-only fields -->

    <template v-if="mode === 'create'">
      <!-- Name -->

      <div class="space-y-2">
        <Label for="name"> Instance name </Label>

        <Input
          id="name"
          v-model="form.name"
          name="name"
          placeholder="my-postgres-db"
          autocomplete="off"
          :aria-invalid="!!errors.name"
        />

        <p v-if="errors.name" class="text-sm text-destructive">
          {{ errors.name }}
        </p>
      </div>

      <!-- Instances -->

      <div class="space-y-2">
        <Label for="instances"> PostgreSQL instances </Label>

        <Input
          id="instances"
          v-model.number="form.instances"
          name="instances"
          type="number"
          min="1"
          max="20"
          :aria-invalid="!!errors.instances"
        />

        <p v-if="errors.instances" class="text-sm text-destructive">
          {{ errors.instances }}
        </p>
      </div>
    </template>

    <!-- Version -->

    <div class="space-y-2">
      <Label for="version"> PostgreSQL version </Label>

      <Input
        id="version"
        v-model.number="form.version"
        name="version"
        type="number"
        min="14"
        max="16"
        :aria-invalid="!!errors.version"
      />

      <p v-if="errors.version" class="text-sm text-destructive">
        {{ errors.version }}
      </p>
    </div>

    <!-- Storage -->

    <div class="space-y-2">
      <Label for="storage"> Storage </Label>

      <Input
        id="storage"
        v-model="form.storage"
        name="storage"
        placeholder="10Gi"
        :aria-invalid="!!errors.storage"
      />

      <p v-if="errors.storage" class="text-sm text-destructive">
        {{ errors.storage }}
      </p>
    </div>

    <Button type="submit" class="w-full" :disabled="loading">
      {{ loading ? 'Saving...' : mode === 'edit' ? 'Save changes' : 'Create instance' }}
    </Button>
  </form>
</template>
