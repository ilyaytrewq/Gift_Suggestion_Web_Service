import type { PropsWithChildren } from 'react';
import {
  useCallback,
  useEffect,
  useMemo,
  useState,
} from 'react';

import type { UserProfile } from '../api/contracts';
import { authSession } from './session';
import type {
  AuthContextValue,
  AuthStatus,
} from './auth-context-object';
import { AuthContext } from './auth-context-object';

export function AuthProvider({ children }: PropsWithChildren): JSX.Element {
  const [accessToken, setAccessToken] = useState<string | null>(
    authSession.getAccessToken(),
  );
  const [user, setUser] = useState<UserProfile | null>(null);
  const [status, setStatus] = useState<AuthStatus>('bootstrapping');

  useEffect(() => authSession.subscribe(() => {
    setAccessToken(authSession.getAccessToken());
  }), []);

  const clearSession = useCallback(() => {
    authSession.clearAccessToken();
    setUser(null);
  }, []);

  const setSession = useCallback((nextAccessToken: string, nextUser: UserProfile | null) => {
    authSession.setAccessToken(nextAccessToken);
    setUser(nextUser);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      accessToken,
      status,
      user,
      clearSession,
      setBootstrapStatus: setStatus,
      setCurrentUser: setUser,
      setSession,
    }),
    [accessToken, clearSession, setSession, status, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
