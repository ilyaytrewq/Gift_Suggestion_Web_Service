import { authSession } from '../auth/session';
import { API_BASE_URL } from '../config/env';
import type { RefreshResponse } from './contracts';
import { ApiError, parseApiError } from './api-error';

type RequestOptions = Omit<RequestInit, 'body'> & {
  auth?: boolean;
  retryOnUnauthorized?: boolean;
  body?: BodyInit | FormData | Record<string, unknown> | unknown[] | null;
};

function isNativeBody(value: RequestOptions['body']): value is BodyInit | FormData {
  return (
    value instanceof FormData ||
    value instanceof Blob ||
    value instanceof URLSearchParams ||
    typeof value === 'string' ||
    value instanceof ArrayBuffer ||
    ArrayBuffer.isView(value)
  );
}

async function refreshAccessToken(): Promise<string | null> {
  const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
  });

  if (!response.ok) {
    authSession.clearAccessToken();
    return null;
  }

  const payload = (await response.json()) as RefreshResponse;
  const nextToken = payload.data.auth.access_token;

  authSession.setAccessToken(nextToken);

  return nextToken;
}

async function performRequest<T>(
  path: string,
  options: RequestOptions = {},
): Promise<T> {
  const {
    auth = false,
    retryOnUnauthorized = true,
    headers,
    body,
    ...init
  } = options;

  const requestHeaders = new Headers(headers);
  let requestBody: BodyInit | null | undefined;

  if (body === undefined || body === null) {
    requestBody = body;
  } else if (isNativeBody(body)) {
    requestBody = body;
  } else {
    requestHeaders.set('Content-Type', 'application/json');
    requestBody = JSON.stringify(body);
  }

  if (auth) {
    const token = authSession.getAccessToken();
    if (token) {
      requestHeaders.set('Authorization', `Bearer ${token}`);
    }
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    body: requestBody,
    headers: requestHeaders,
    credentials: 'include',
  });

  if (
    response.status === 401 &&
    auth &&
    retryOnUnauthorized
  ) {
    const nextToken = await refreshAccessToken();

    if (nextToken) {
      return performRequest<T>(path, {
        ...options,
        retryOnUnauthorized: false,
      });
    }
  }

  if (!response.ok) {
    throw await parseApiError(response);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return (await response.json()) as T;
}

export async function requestJson<T>(
  path: string,
  options?: RequestOptions,
): Promise<T> {
  return performRequest<T>(path, options);
}

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError;
}
