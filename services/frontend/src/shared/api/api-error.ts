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

export function getUserFacingApiErrorMessage(error: ApiError): string {
  switch (error.code) {
    case 'invalid_credentials':
      return 'Почта или пароль указаны неверно.';
    case 'email_already_exists':
      return 'Этот email уже используется.';
    case 'invalid_email':
      return 'Проверьте email и попробуйте снова.';
    case 'invalid_password':
      return 'Введите пароль.';
    case 'missing_access_token':
    case 'invalid_refresh_token':
      return 'Войдите, чтобы продолжить.';
    case 'user_not_found':
      return 'Не удалось найти данные аккаунта.';
    case 'gift_not_found':
      return 'Подарок не найден.';
    case 'invalid_display_name':
      return 'Проверьте имя и попробуйте снова.';
    case 'invalid_profile_update':
      return 'Укажите, что хотите изменить.';
    case 'budget_required':
      return 'Укажите бюджет.';
    case 'invalid_budget':
      return 'Проверьте бюджет и попробуйте снова.';
    case 'invalid_occasion':
      return 'Проверьте повод.';
    case 'invalid_relationship':
      return 'Проверьте, кому вы выбираете подарок.';
    case 'invalid_recipient_age':
      return 'Проверьте возраст получателя.';
    case 'invalid_top_n':
      return 'Выберите количество вариантов от 1 до 20.';
    case 'invalid_interests':
      return 'Проверьте список интересов.';
    case 'invalid_recommendation_request':
      return 'Не удалось обработать запрос. Проверьте данные и попробуйте ещё раз.';
    default:
      break;
  }

  switch (error.status) {
    case 401:
    case 403:
      return 'Войдите, чтобы продолжить.';
    case 404:
      return 'Ничего не найдено.';
    case 409:
      return 'Такое действие сейчас недоступно. Проверьте данные и попробуйте снова.';
    case 503:
      return 'Сервис временно недоступен. Попробуйте позже.';
    default:
      return 'Что-то пошло не так. Попробуйте ещё раз.';
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
