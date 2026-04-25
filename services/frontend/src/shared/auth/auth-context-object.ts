import { createContext } from 'react';

import type { UserProfile } from '../api/contracts';

export type AuthStatus = 'bootstrapping' | 'ready';

export type AuthContextValue = {
  accessToken: string | null;
  status: AuthStatus;
  user: UserProfile | null;
  clearSession: () => void;
  setBootstrapStatus: (status: AuthStatus) => void;
  setCurrentUser: (user: UserProfile | null) => void;
  setSession: (accessToken: string, user: UserProfile | null) => void;
};

export const AuthContext = createContext<AuthContextValue | null>(null);
