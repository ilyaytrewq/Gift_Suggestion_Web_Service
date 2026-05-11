import { useQueryClient } from '@tanstack/react-query';
import { useEffect, useRef, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { exchangeVkOAuth } from '../../../features/vk-integration/api/vk';
import {
  parseVkOAuthCallback,
  VK_OAUTH_CODE_VERIFIER_KEY,
  VK_OAUTH_PENDING_KEY,
  VK_OAUTH_STATE_KEY,
  type VkOAuthCallbackSuccess,
} from '../../../features/vk-integration/lib/vk-oauth';
import { VK_CONSENT_POLICY_VERSION } from '../../../features/vk-integration/model/constants';
import { useAuth } from '../../../shared/auth/use-auth';
import { isVkOAuthConfigured, getVkOAuthRedirectUri } from '../../../shared/config/env';
import { getUserFacingApiErrorMessage } from '../../../shared/api/api-error';
import { isApiError } from '../../../shared/api/http';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { PageLoader } from '../../../shared/ui/feedback/page-loader';
import { Container } from '../../../shared/ui/layout/container';

function exchangePayloadFromCallback(data: VkOAuthCallbackSuccess, codeVerifier: string) {
  return {
    code: data.code,
    code_verifier: codeVerifier,
    device_id: data.device_id,
    ...(data.state ? { state: data.state } : {}),
    ...(getVkOAuthRedirectUri() ? { redirect_uri: getVkOAuthRedirectUri() } : {}),
    consent: {
      granted: true as const,
      version: VK_CONSENT_POLICY_VERSION,
      obtained_at: new Date().toISOString(),
    },
  };
}

export function VkOAuthCallbackPage(): JSX.Element {
  const auth = useAuth();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const capturedSearchRef = useRef(
    typeof window !== 'undefined' ? window.location.search : '',
  );
  const [error, setError] = useState<string | null>(null);
  const [done, setDone] = useState(false);

  useEffect(() => {
    let cancelled = false;

    async function run() {
      if (!isVkOAuthConfigured()) {
        navigate('/profile', { replace: true });
        return;
      }

      const pendingRaw = sessionStorage.getItem(VK_OAUTH_PENDING_KEY);
      const codeVerifier = sessionStorage.getItem(VK_OAUTH_CODE_VERIFIER_KEY);

      if (auth.accessToken && pendingRaw && codeVerifier) {
        try {
          const pending = JSON.parse(pendingRaw) as VkOAuthCallbackSuccess;
          sessionStorage.removeItem(VK_OAUTH_PENDING_KEY);
          sessionStorage.removeItem(VK_OAUTH_CODE_VERIFIER_KEY);
          await exchangeVkOAuth(exchangePayloadFromCallback(pending, codeVerifier));
          if (cancelled) {
            return;
          }
          await queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
          navigate('/profile', { replace: true });
          return;
        } catch (unknownErr) {
          sessionStorage.removeItem(VK_OAUTH_PENDING_KEY);
          sessionStorage.removeItem(VK_OAUTH_CODE_VERIFIER_KEY);
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

      const search = capturedSearchRef.current;
      if (!search || search === '?') {
        if (auth.accessToken) {
          navigate('/profile', { replace: true });
          return;
        }
        if (!cancelled) {
          setDone(true);
        }
        return;
      }

      const parsed = parseVkOAuthCallback(search);

      if (!parsed.ok) {
        window.history.replaceState(null, '', window.location.pathname);
        if (!cancelled) {
          setError(parsed.reason);
          setDone(true);
        }
        return;
      }

      const savedState = sessionStorage.getItem(VK_OAUTH_STATE_KEY);
      sessionStorage.removeItem(VK_OAUTH_STATE_KEY);
      if (parsed.data.state && savedState && parsed.data.state !== savedState) {
        window.history.replaceState(null, '', window.location.pathname);
        if (!cancelled) {
          setError(
            'Параметр безопасности state не совпадает. Повторите вход через VK.',
          );
          setDone(true);
        }
        return;
      }

      const verifier = sessionStorage.getItem(VK_OAUTH_CODE_VERIFIER_KEY);
      if (!verifier) {
        window.history.replaceState(null, '', window.location.pathname);
        if (!cancelled) {
          setError('Сессия VK ID истекла. Начните подключение заново из профиля.');
          setDone(true);
        }
        return;
      }

      if (!auth.accessToken) {
        sessionStorage.setItem(VK_OAUTH_PENDING_KEY, JSON.stringify(parsed.data));
        window.history.replaceState(null, '', window.location.pathname);
        navigate(`/login?next=${encodeURIComponent('/auth/vk-callback')}`, {
          replace: true,
        });
        return;
      }

      try {
        await exchangeVkOAuth(exchangePayloadFromCallback(parsed.data, verifier));
        sessionStorage.removeItem(VK_OAUTH_CODE_VERIFIER_KEY);
        if (cancelled) {
          return;
        }
        window.history.replaceState(null, '', window.location.pathname);
        await queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
        navigate('/profile', { replace: true });
      } catch (unknownErr) {
        sessionStorage.removeItem(VK_OAUTH_CODE_VERIFIER_KEY);
        window.history.replaceState(null, '', window.location.pathname);
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
        description="Обмениваем код VK ID на токен…"
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
