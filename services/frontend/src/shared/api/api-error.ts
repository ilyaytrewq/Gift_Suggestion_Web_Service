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
    case 'email_not_verified':
      return 'Подтвердите email: откройте ссылку из письма или введите код на странице подтверждения, затем войдите снова.';
    case 'email_already_exists':
      return 'Этот email уже используется.';
    case 'invalid_email':
      return 'Проверьте email и попробуйте снова.';
    case 'invalid_password':
      return 'Пароль не соответствует требованиям безопасности.';
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
    case 'vk_connection_already_exists':
      return 'Уже привязан другой аккаунт VK. Сначала отключите текущий.';
    case 'vk_token_storage_not_configured':
    case 'vk_token_storage_unavailable':
      return 'На сервере не настроено хранение токена VK (ключ шифрования).';
    case 'vk_consent_required':
      return 'Нужно ваше согласие на обработку данных VK.';
    case 'vk_token_expired':
      return 'Срок действия токена VK истёк. Подключите аккаунт снова.';
    case 'vk_token_invalid':
      return 'Токен VK недействителен. Отключите и снова подключите аккаунт VK.';
    case 'vk_connection_not_ready':
      return 'Подключите VK и сохраните токен, затем повторите синхронизацию.';
    case 'vk_rate_limited':
      return 'VK API временно ограничил запросы. Повторите через несколько секунд.';
    case 'vk_groups_access_denied':
      return 'Список групп VK закрыт настройками приватности. Откройте доступ к группам в настройках VK и повторите.';
    case 'vk_groups_scope_required':
      return 'Импорт групп недоступен для текущего токена. Синхронизация профиля VK ID должна пройти автоматически; для групп запросите право groups в id.vk.com.';
    case 'vk_interest_import_timeout':
      return 'VK API не ответил вовремя. Повторите синхронизацию позже.';
    case 'vk_interest_import_unavailable':
      return 'Импорт интересов из VK недоступен. Обратитесь к администратору.';
    case 'vk_integration_disabled':
      return 'Интеграция VK на сервере отключена.';
    case 'invalid_content_type':
      return 'Неверный тип запроса: нужно отправить файл как multipart/form-data.';
    case 'file_too_large':
      return 'Файл импорта слишком большой. Уменьшите размер или увеличьте лимит на сервере.';
    case 'file_required':
      return 'Не выбран файл. Укажите CSV, JSON или XLSX.';
    case 'unsupported_import_format':
      return 'Поддерживаются только расширения файла .csv, .json и .xlsx.';
    case 'empty_import_file':
      return 'Файл пуст или не содержит записей каталога.';
    case 'import_file_read_failed':
      return 'Не удалось прочитать загруженный файл. Попробуйте другой файл или повторите позже.';
    case 'invalid_import_job_id':
    case 'import_job_not_found':
      return 'Задача импорта не найдена. Обновите страницу и попробуйте снова.';
    case 'invalid_vk_connection_payload':
      return 'Проверьте данные подключения VK и попробуйте снова.';
    default:
      break;
  }

  const serverMsg = error.message.trim();
  if (
    serverMsg &&
    !serverMsg.startsWith('Request failed with status')
  ) {
    return serverMsg;
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
