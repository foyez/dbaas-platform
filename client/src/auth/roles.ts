import type { User } from 'oidc-client-ts'

const ZITADEL_ROLES_CLAIM = 'urn:zitadel:iam:org:project:roles'

export function hasRole(user: User | null, role: string): boolean {
  if (!user) return false

  const roles = user.profile[ZITADEL_ROLES_CLAIM] as Record<string, unknown> | undefined
  return !!roles && role in roles
}