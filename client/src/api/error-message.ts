import type { ApiError } from './errors'

export function getInstanceErrorMessage(error: ApiError): string {
  switch (error.code) {
    case 'INSTANCE_NAME_CONFLICT':
      return 'An instance with this name already exists.'

    case 'INVALID_INSTANCE_CONFIGURATION':
      return error.message

    case 'UNAUTHENTICATED':
      return 'Please sign in again.'

    default:
      return 'Unable to create the instance. Please try again.'
  }
}
