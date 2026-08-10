import type { InstanceStatus } from '@/types/instance'

export function getInstanceStatusDescription(status: InstanceStatus): string {
  switch (status) {
    case 'Pending':
      return 'Your instance has been accepted and provisioning will begin shortly.'

    case 'Provisioning':
      return 'Your PostgreSQL instance is currently being provisioned.'

    case 'Running':
      return 'Your PostgreSQL instance is running normally.'

    case 'Failed':
      return 'Provisioning failed. Please check the instance details or try again.'

    case 'Deleting':
      return 'Your PostgreSQL instance is being deleted.'

    case 'Deleted':
      return 'This instance has been deleted.'

    default:
      return ''
  }
}
