import type { PropsWithChildren } from 'react';
import { useEffect } from 'react';

import { getCurrentUser, refreshToken } from '../../features/auth/api/auth';
import { PageLoader } from '../ui/feedback/page-loader';
import { useAuth } from './use-auth';

export function AuthBootstrap({ children }: PropsWithChildren): JSX.Element {
  const {
    clearSession,
    setBootstrapStatus,
    setCurrentUser,
    setSession,
    status,
  } = useAuth();

  useEffect(() => {
    let cancelled = false;

    async function bootstrap(): Promise<void> {
      try {
        const refreshPayload = await refreshToken();

        if (cancelled) {
          return;
        }

        setSession(refreshPayload.data.auth.access_token, null);

        const currentUser = await getCurrentUser();

        if (cancelled) {
          return;
        }

        setCurrentUser(currentUser.data.user);
      } catch {
        if (!cancelled) {
          clearSession();
        }
      } finally {
        if (!cancelled) {
          setBootstrapStatus('ready');
        }
      }
    }

    void bootstrap();

    return () => {
      cancelled = true;
    };
  }, [clearSession, setBootstrapStatus, setCurrentUser, setSession]);

  if (status === 'bootstrapping') {
    return (
      <PageLoader
        title="Подготавливаем приложение"
        description="Проверяем активную сессию и подгружаем базовый контекст."
      />
    );
  }

  return <>{children}</>;
}
