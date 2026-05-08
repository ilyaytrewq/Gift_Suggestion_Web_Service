const configuredApiBaseUrl = import.meta.env.VITE_API_BASE_URL;

export const API_BASE_URL =
  configuredApiBaseUrl === undefined
    ? 'http://localhost:8080'
    : configuredApiBaseUrl.trim().replace(/\/$/, '');

/** Implicit OAuth; redirect URI must be listed in VK app settings. */
export function isVkOAuthConfigured(): boolean {
  const clientId = import.meta.env.VITE_VK_APP_ID;
  const redirectUri = import.meta.env.VITE_VK_REDIRECT_URI;
  return (
    typeof clientId === 'string' &&
    clientId.trim().length > 0 &&
    typeof redirectUri === 'string' &&
    redirectUri.trim().length > 0
  );
}

export function getVkOAuthClientId(): string | undefined {
  const clientId = import.meta.env.VITE_VK_APP_ID;
  return typeof clientId === 'string' && clientId.trim().length > 0
    ? clientId.trim()
    : undefined;
}

export function getVkOAuthRedirectUri(): string | undefined {
  const redirectUri = import.meta.env.VITE_VK_REDIRECT_URI;
  return typeof redirectUri === 'string' && redirectUri.trim().length > 0
    ? redirectUri.trim()
    : undefined;
}
