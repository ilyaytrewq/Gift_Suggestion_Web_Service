import { getVkOAuthClientId, getVkOAuthRedirectUri } from '../../../shared/config/env';

export const VK_OAUTH_SCOPE = 'friends,groups';

/** SessionStorage marker for CSRF on implicit redirect. */
export const VK_OAUTH_STATE_KEY = 'vk_oauth_state';

/** Saved OAuth payload when user must log in before calling `connectVk`. */
export const VK_OAUTH_PENDING_KEY = 'vk_oauth_pending_connect';

export function generateVkOAuthState(): string {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }
  return `vk_${Date.now()}_${Math.random().toString(36).slice(2)}`;
}

export function buildVkImplicitAuthorizeUrl(state: string): string {
  const clientId = getVkOAuthClientId();
  const redirectUri = getVkOAuthRedirectUri();
  if (!clientId || !redirectUri) {
    throw new Error('VK OAuth is not configured');
  }

  const params = new URLSearchParams({
    client_id: clientId,
    display: 'page',
    redirect_uri: redirectUri,
    scope: VK_OAUTH_SCOPE,
    response_type: 'token',
    v: '5.199',
    state,
  });

  return `https://oauth.vk.com/authorize?${params.toString()}`;
}

export type VkOAuthFragmentSuccess = {
  access_token: string;
  user_id: string;
  expires_in: number | undefined;
  scope: string[];
  state?: string | null;
};

export type VkOAuthFragmentResult =
  | { ok: true; data: VkOAuthFragmentSuccess }
  | { ok: false; reason: string };

function splitScopes(fragmentValue: string | null): string[] {
  if (!fragmentValue) {
    return [];
  }
  return [...new Set(
    fragmentValue
      .split(/[,+]/)
      .map((segment) => segment.trim().toLowerCase())
      .filter(Boolean),
  )];
}

/**
 * Parses the fragment returned by oauth.vk.com for implicit (response_type=token) flow.
 */
export function parseVkOAuthFragment(fragment: string): VkOAuthFragmentResult {
  const normalized = fragment.startsWith('#')
    ? fragment.slice(1)
    : fragment;
  const params = new URLSearchParams(normalized);
  const error = params.get('error');
  const errorDescription = params.get('error_description');

  if (error || errorDescription) {
    const message = [errorDescription, error].filter(Boolean).join(' ');
    return { ok: false, reason: message || 'Авторизация VK отклонена' };
  }

  const accessToken = params.get('access_token');
  const userId = params.get('user_id');
  if (!accessToken || !userId) {
    return {
      ok: false,
      reason: 'Не удалось получить токен VK. Проверьте параметры приложения и redirect_uri.',
    };
  }

  const expiresInRaw = params.get('expires_in');
  const expiresInParsed = expiresInRaw ? Number.parseInt(expiresInRaw, 10) : undefined;
  const expiresIn =
    expiresInParsed !== undefined && Number.isFinite(expiresInParsed) && expiresInParsed > 0
      ? expiresInParsed
      : undefined;

  return {
    ok: true,
    data: {
      access_token: accessToken,
      user_id: userId,
      expires_in: expiresIn,
      scope: splitScopes(params.get('scope')),
      state: params.get('state'),
    },
  };
}
