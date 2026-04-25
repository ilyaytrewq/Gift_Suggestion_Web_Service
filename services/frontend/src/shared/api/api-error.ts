import type { ErrorEnvelope } from './contracts';

export class ApiError extends Error {
  public readonly status: number;

  public readonly code?: string;

  public readonly requestId?: string;

  constructor(
    message: string,
    options: { status: number; code?: string; requestId?: string },
  ) {
    super(message);
    this.name = 'ApiError';
    this.status = options.status;
    this.code = options.code;
    this.requestId = options.requestId;
  }
}

export async function parseApiError(response: Response): Promise<ApiError> {
  let payload: ErrorEnvelope | null = null;

  try {
    payload = (await response.json()) as ErrorEnvelope;
  } catch {
    payload = null;
  }

  return new ApiError(
    payload?.error.message ?? `Request failed with status ${response.status}`,
    {
      status: response.status,
      code: payload?.error.code,
      requestId: payload?.meta.request_id,
    },
  );
}
