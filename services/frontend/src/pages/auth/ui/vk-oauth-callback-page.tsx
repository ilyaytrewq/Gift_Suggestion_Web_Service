import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { connectVk } from '../../../features/vk-integration/api/vk';
import {
  parseVkOAuthFragment,
  VK_OAUTH_PENDING_KEY,
  VK_OAUTH_STATE_KEY,
  type VkOAuthFragmentSuccess,
} from '../../../features/vk-integration/lib/vk-oauth';
import { VK_CONSENT_POLICY_VERSION } from '../../../features/vk-integration/model/constants';
import { useAuth } from '../../../shared/auth/use-auth';
import { getUserFacingApiErrorMessage } from '../../../shared/api/api-error';
import { isApiError } from '../../../shared/api/http';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { PageLoader } from '../../../shared/ui/feedback/page-loader';
import { Container } from '../../../shared/ui/layout/container';

function connectPayloadFromOAuth(data: VkOAuthFragmentSuccess) {
  const expiresAtIso =
    data.expires_in !== undefined
      ? new Date(Date.now() + data.expires_in * 1000).toISOString()
      : undefined;

  return {
    provider_user_id: data.user_id,
    consent: {
      granted: true as const,
      version: VK_CONSENT_POLICY_VERSION,
      obtained_at: new Date().toISOString(),
    },
    credential: {
      access_token: data.access_token,
      ...(expiresAtIso ? { expires_at: expiresAtIso } : {}),
      ...(data.scope.length ? { scopes: data.scope } : {}),
    },
  };
}

export function VkOAuthCallbackPage(): JSX.Element {
  const auth = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  /** Captured once per mount so StrictMode / replaceState does not lose the fragment. */
  const capturedHashRef = useRef(
    typeof window !== 'undefined' ? window.location.hash : '',
  );
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      const pendingRaw = sessionStorage.getItem(VK_OAUTH_PENDING_KEY);

      if (auth.accessToken && pendingRaw) {
        try {
          const pending = JSON.parse(pendingRaw) as VkOAuthFragmentSuccess;
          sessionStorage.removeItem(VK_OAUTH_PENDING_KEY);
          await connectVk(connectPayloadFromOAuth(pending));
          if (cancelled) {
            return;
          }
          await queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
          navigate('/profile', { replace: true });
          return;
        } catch (unknownErr) {
          sessionStorage.removeItem(VK_OAUTH_PENDING_KEY);
          if (!cancelled) {
            setError(
              isApiError(unknownErr)
                ? getUserFacingApiErrorMessage(unknownErr)
                : unknownErr instanceof Error
                  ? unknownErr.message
                  : 'Не удалось подключить VK.',
            );
            setDone(true);
          }
          return;
        }
      }

      const fragment = capturedHashRef.current;
      if (!fragment || fragment === '#') {
        if (auth.accessToken) {
          navigate('/profile', { replace: true });
          return;
        }
        if (!cancelled) {
          setDone(true);
        }
        return;
      }

      const parsed = parseVkOAuthFragment(fragment);

      if (!parsed.ok) {
        window.history.replaceState(
          null,
          '',
          `${window.location.pathname}${window.location.search}`,
        );
        if (!cancelled) {
          setError(parsed.reason);
          setDone(true);
        }
        return;
      }

      const savedState = sessionStorage.getItem(VK_OAUTH_STATE_KEY);
      sessionStorage.removeItem(VK_OAUTH_STATE_KEY);
      if (
        parsed.data.state &&
        savedState &&
        parsed.data.state !== savedState
      ) {
        window.history.replaceState(
          null,
          '',
          `${window.location.pathname}${window.location.search}`,
        );
        if (!cancelled) {
          setError(
            'Параметр безопасности state не совпадает. Повторите вход через VK.',
          );
          setDone(true);
        }
        return;
      }

      if (!auth.accessToken) {
        sessionStorage.setItem(
          VK_OAUTH_PENDING_KEY,
          JSON.stringify(parsed.data),
        );
        window.history.replaceState(
          null,
          '',
          `${window.location.pathname}${window.location.search}`,
        );
        navigate(
          `/login?next=${encodeURIComponent('/auth/vk-callback')}`,
          { replace: true },
        );
        return;
      }

      try {
        await connectVk(connectPayloadFromOAuth(parsed.data));
        if (cancelled) {
          return;
        }
        window.history.replaceState(
          null,
          '',
          `${window.location.pathname}${window.location.search}`,
        );
        await queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
        navigate('/profile', { replace: true });
      } catch (unknownErr) {
        window.history.replaceState(
          null,
          '',
          `${window.location.pathname}${window.location.search}`,
        );
        if (!cancelled) {
          setError(
            isApiError(unknownErr)
              ? getUserFacingApiErrorMessage(unknownErr)
              : unknownErr instanceof Error
                ? unknownErr.message
                : 'Не удалось подключить VK.',
          );
          setDone(true);
        }
      }
    }

    void run();

    return () => {
      cancelled = true;
    };
  }, [auth.accessToken, navigate, queryClient]);

  if (!done && !error) {
    return (
      <PageLoader
        description="Сохраняем токен на сервере…"
        title="Подключаем VK"
      />
    );
  }

  if (error) {
    return (
      <Container className="page-stack">
        <section className="auth-form">
          <p className="eyebrow">VK</p>
          <h1>Не удалось завершить подключение</h1>
          <p className="page-copy">{error}</p>
          <Link className={buttonClassName()} to="/profile">
            В профиль
          </Link>
        </section>
      </Container>
    );
  }

  return (
    <Container className="page-stack">
      <section className="auth-form">
        <p className="eyebrow">VK</p>
        <h1>Ожидаем данные от VK</h1>
        <p className="page-copy">
          Откройте эту страницу из потока «Войти через VK» на странице профиля или войдите в аккаунт
          и повторите подключение.
        </p>
        <Link className={buttonClassName()} to="/login">
          Войти
        </Link>{' '}
        <Link className={buttonClassName({ variant: 'ghost' })} to="/profile">
          В профиль
        </Link>
      </section>
    </Container>
  );
}
