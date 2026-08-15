import * as z from 'zod'

export const createInstanceSchema = z.object({
  name: z
    .string()
    .trim()
    .superRefine((value, ctx) => {
      if (!value) {
        ctx.addIssue({
          code: 'custom',
          message: 'Name is required',
        })
        return
      }

      if (value.length > 63) {
        ctx.addIssue({
          code: 'custom',
          message: 'Name must be 63 characters or fewer',
        })
        return
      }

      if (!/^[a-z0-9](?:[-a-z0-9]*[a-z0-9])?$/.test(value)) {
        ctx.addIssue({
          code: 'custom',
          message: 'Use lowercase letters, numbers, and hyphens only',
        })
      }
    }),

  instances: z.preprocess(
    (value) => (value === '' ? 0 : value),
    z
      .number()
      .int('Instances must be a whole number')
      .min(1, 'At least 1 instance is required')
      .max(20, 'Maximum 20 instances'),
  ),

  version: z.preprocess(
    (value) => (value === '' ? 0 : value),
    z
      .number()
      .int('Version must be a whole number')
      .min(12, 'Minimum PostgreSQL version is 12')
      .max(18, 'Maximum PostgreSQL version is 18'),
  ),

  storage: z
    .string()
    .trim()
    .regex(/^\d+(Mi|Gi|Ti)$/, 'Storage must look like 512Mi, 7Gi, or 1Ti'),
})

export const updateInstanceSchema = createInstanceSchema.pick({
  version: true,
  storage: true,
})

export type CreateInstanceInput = z.infer<typeof createInstanceSchema>
export type UpdateInstanceInput = z.infer<typeof updateInstanceSchema>
