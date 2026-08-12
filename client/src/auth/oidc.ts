import { UserManager, type UserManagerSettings } from 'oidc-client-ts'

export const oidcSettings: UserManagerSettings = {
  authority: import.meta.env.VITE_ZITADEL_AUTHORITY,
  client_id: import.meta.env.VITE_ZITADEL_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_ZITADEL_REDIRECT_URI,
  post_logout_redirect_uri: import.meta.env.VITE_ZITADEL_POST_LOGOUT_URI,

  response_type: 'code',
  scope: 'openid profile email offline_access',

  // Renew using the refresh token when available.
  // automaticSilentRenew: true,
  loadUserInfo: true,
}

// Read-only handle used outside components (e.g. api/client.ts).
// Same storage key as AuthProvider's internal UserManager, so both
// stay in sync without needing Vue's inject() context.
export const userManager = new UserManager(oidcSettings)

export function onSigninCallback() {
  // Remove ?code=...&state=... from the browser URL
  const url = new URL(window.location.href)

  url.search = ''
  url.hash = ''

  window.history.replaceState({}, document.title, window.location.pathname)
}
