import { getVkOAuthClientId, getVkOAuthRedirectUri } from '../../../shared/config/env';

/** VK ID scopes (space-separated in authorize URL). */
export const VK_OAUTH_SCOPE = 'groups';

const VK_ID_AUTHORIZE_URL = 'https://id.vk.ru/authorize';

/** SessionStorage marker for CSRF on VK ID redirect. */
export const VK_OAUTH_STATE_KEY = 'vk_oauth_state';

/** PKCE code_verifier for the redirect round-trip. */
export const VK_OAUTH_CODE_VERIFIER_KEY = 'vk_oauth_code_verifier';

/** Saved OAuth payload when user must log in before completing exchange. */
export const VK_OAUTH_PENDING_KEY = 'vk_oauth_pending_connect';

const CODE_VERIFIER_CHARSET =
  'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789_-';

export function generateVkOAuthState(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }
  return `vk_${Date.now()}_${Math.random().toString(36).slice(2)}`;
}

export function generateVkOAuthCodeVerifier(): string {
  const length = 64;
  if (typeof crypto !== 'undefined' && crypto.getRandomValues) {
    const bytes = new Uint8Array(length);
    crypto.getRandomValues(bytes);
    return Array.from(bytes, (byte) => CODE_VERIFIER_CHARSET[byte % CODE_VERIFIER_CHARSET.length]).join('');
  }
  return Array.from({ length }, () =>
    CODE_VERIFIER_CHARSET[Math.floor(Math.random() * CODE_VERIFIER_CHARSET.length)],
  ).join('');
}

export async function computeVkOAuthCodeChallenge(codeVerifier: string): Promise<string> {
  const data = new TextEncoder().encode(codeVerifier);
  const digest = await crypto.subtle.digest('SHA-256', data);
  return base64UrlEncode(new Uint8Array(digest));
}

function base64UrlEncode(bytes: Uint8Array): string {
  let binary = '';
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/u, '');
}

export function buildVkIdAuthorizeUrl(state: string, codeChallenge: string): string {
  const clientId = getVkOAuthClientId();
  const redirectUri = getVkOAuthRedirectUri();
  if (!clientId || !redirectUri) {
    throw new Error('VK OAuth is not configured');
  }

  const params = new URLSearchParams({
    response_type: 'code',
    client_id: clientId,
    redirect_uri: redirectUri,
    scope: VK_OAUTH_SCOPE,
    state,
    code_challenge: codeChallenge,
    code_challenge_method: 'S256',
  });

  return `${VK_ID_AUTHORIZE_URL}?${params.toString()}`;
}

export type VkOAuthCallbackSuccess = {
  code: string;
  device_id: string;
  state: string;
};

export type VkOAuthCallbackResult =
  | { ok: true; data: VkOAuthCallbackSuccess }
  | { ok: false; reason: string };

/**
 * Parses query parameters returned by id.vk.ru after authorization.
 */
export function parseVkOAuthCallback(search: string): VkOAuthCallbackResult {
  const params = new URLSearchParams(search.startsWith('?') ? search.slice(1) : search);
  const error = params.get('error');
  const errorDescription = params.get('error_description');

  if (error || errorDescription) {
    const message = [errorDescription, error].filter(Boolean).join(' ');
    return { ok: false, reason: message || 'Авторизация VK отклонена' };
  }

  const code = params.get('code');
  const deviceId = params.get('device_id');
  const state = params.get('state');
  if (!code || !deviceId) {
    return {
      ok: false,
      reason:
        'Не удалось получить код VK ID. Проверьте redirect URI в настройках приложения id.vk.com.',
    };
  }
  if (!state) {
    return {
      ok: false,
      reason: 'VK ID не вернул параметр state. Повторите вход через VK.',
    };
  }

  return {
    ok: true,
    data: {
      code,
      device_id: deviceId,
      state,
    },
  };
}
