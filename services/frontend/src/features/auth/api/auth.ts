import { authSession } from '../../../shared/auth/session';
import type {
  CurrentUserResponse,
  LoginRequest,
  LoginResponse,
  LogoutResponse,
  PasswordResetRequest,
  PasswordResetResponse,
  RegisterRequest,
  RegisterResponse,
  RefreshResponse,
  UpdateCurrentUserRequest,
  UpdateCurrentUserResponse,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';

export function loginUser(payload: LoginRequest): Promise<LoginResponse> {
  return requestJson<LoginResponse>('/api/v1/auth/login', {
    method: 'POST',
    body: payload,
  });
}

export function registerUser(payload: RegisterRequest): Promise<RegisterResponse> {
  return requestJson<RegisterResponse>('/api/v1/users', {
    method: 'POST',
    body: payload,
  });
}

export function requestPasswordReset(
  payload: PasswordResetRequest,
): Promise<PasswordResetResponse> {
  return requestJson<PasswordResetResponse>('/api/v1/auth/password-reset/request', {
    method: 'POST',
    body: payload,
  });
}

export function refreshToken(): Promise<RefreshResponse> {
  return requestJson<RefreshResponse>('/api/v1/auth/refresh', {
    method: 'POST',
    retryOnUnauthorized: false,
  });
}

export function logoutUser(): Promise<LogoutResponse> {
  return requestJson<LogoutResponse>('/api/v1/auth/logout', {
    method: 'POST',
    retryOnUnauthorized: false,
  });
}

export function getCurrentUser(): Promise<CurrentUserResponse> {
  return requestJson<CurrentUserResponse>('/api/v1/users/me', {
    auth: true,
  });
}

export function updateCurrentUser(
  payload: UpdateCurrentUserRequest,
): Promise<UpdateCurrentUserResponse> {
  return requestJson<UpdateCurrentUserResponse>('/api/v1/users/me', {
    method: 'PATCH',
    auth: true,
    body: payload,
  });
}

export function applyAuthenticatedSession(
  accessToken: string,
): void {
  authSession.setAccessToken(accessToken);
}
