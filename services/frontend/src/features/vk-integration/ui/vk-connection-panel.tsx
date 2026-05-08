import { zodResolver } from '@hookform/resolvers/zod';
import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';
import { useForm, Controller } from 'react-hook-form';
import { z } from 'zod';

import { connectVk, disconnectVk, getVkConnection, syncVkInterests } from '../api/vk';
import {
  buildVkImplicitAuthorizeUrl,
  generateVkOAuthState,
  VK_OAUTH_STATE_KEY,
} from '../lib/vk-oauth';
import { VK_CONSENT_POLICY_VERSION } from '../model/constants';
import { useAuth } from '../../../shared/auth/use-auth';
import { getUserFacingApiErrorMessage } from '../../../shared/api/api-error';
import { isApiError } from '../../../shared/api/http';
import type { ConnectVKRequestBody } from '../../../shared/api/contracts';
import type { components } from '../../../shared/api/generated/schema';
import { isVkOAuthConfigured } from '../../../shared/config/env';
import { cn } from '../../../shared/lib/cn';
import { formatDateTime } from '../../../shared/lib/format';
import { Button } from '../../../shared/ui/button/button';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';

type VKConnection = components['schemas']['VKConnection'];

const manualSchema = z.object({
  provider_user_id: z
    .string()
    .trim()
    .min(1, 'Укажите идентификатор профиля VK')
    .max(128),
  access_token: z.string(),
  scopes_csv: z.string(),
  expires_at_local: z.string(),
  screen_name: z.string().max(64),
  profile_url: z
    .string()
    .trim()
    .max(512)
    .refine((value) => value === '' || z.string().url().safeParse(value).success, {
      message: 'Укажите корректный URL профиля',
    }),
  consent_acknowledged: z.boolean().refine((value) => value, {
    message: 'Подтвердите согласие, чтобы сохранить подключение',
  }),
});

type ManualSchema = z.infer<typeof manualSchema>;

function parseScopes(csv: string): string[] {
  return [
    ...new Set(
      csv
        .split(',')
        .map((segment) => segment.trim().toLowerCase())
        .filter(Boolean),
    ),
  ];
}

function mapConnectionUiState(state: VKConnection['state']): string {
  switch (state) {
    case 'connected':
      return 'Подключено';
    case 'sync_required':
      return 'Подключено, ждём синхронизации';
    case 'error':
      return 'Подключено, ошибка синхронизации';
    case 'disconnected':
    default:
      return 'Не подключено';
  }
}

function mapSyncUiStatus(status: VKConnection['sync']['last_status']): string {
  switch (status) {
    case 'succeeded':
      return 'Успешно';
    case 'failed':
      return 'Не удалось';
    case 'idle':
    default:
      return 'Ещё не запускали';
  }
}

function buildManualConnectBody(values: ManualSchema): ConnectVKRequestBody {
  const trimmedToken = values.access_token.trim();
  const scopes = parseScopes(values.scopes_csv);
  const expiresIso =
    trimmedToken !== '' &&
    typeof values.expires_at_local === 'string' &&
    values.expires_at_local.trim() !== ''
      ? new Date(values.expires_at_local).toISOString()
      : undefined;

  const credential =
    trimmedToken !== '' || scopes.length > 0 || expiresIso
      ? {
        ...(trimmedToken !== '' ? { access_token: trimmedToken } : {}),
        ...(scopes.length ? { scopes } : {}),
        ...(expiresIso ? { expires_at: expiresIso } : {}),
      }
      : undefined;

  const screenName = values.screen_name.trim();
  const profileURL = values.profile_url.trim();

  const profile =
    screenName !== '' || profileURL !== ''
      ? {
        ...(screenName !== '' ? { screen_name: screenName } : {}),
        ...(profileURL !== '' ? { profile_url: profileURL } : {}),
      }
      : undefined;

  return {
    provider_user_id: values.provider_user_id.trim(),
    consent: {
      granted: true as const,
      version: VK_CONSENT_POLICY_VERSION,
      obtained_at: new Date().toISOString(),
    },
    ...(credential !== undefined ? { credential } : {}),
    ...(profile !== undefined ? { profile } : {}),
  };
}

export function VkConnectionPanel(): JSX.Element | null {
  const auth = useAuth();
  const queryClient = useQueryClient();

  const connectionQuery = useQuery({
    queryKey: ['vk-connection'],
    queryFn: getVkConnection,
    enabled: Boolean(auth.accessToken),
    retry: (_count, error) => !(isApiError(error) && error.status === 503),
  });

  const manualForm = useForm<ManualSchema>({
    resolver: zodResolver(manualSchema),
    defaultValues: {
      provider_user_id: '',
      access_token: '',
      scopes_csv: '',
      expires_at_local: '',
      screen_name: '',
      profile_url: '',
      consent_acknowledged: false,
    },
  });

  const connectMutation = useMutation({
    mutationFn: connectVk,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
    },
  });

  const disconnectMutation = useMutation({
    mutationFn: disconnectVk,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
      connectMutation.reset();
      manualForm.reset();
    },
  });

  const syncMutation = useMutation({
    mutationFn: syncVkInterests,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
    },
  });

  const handleStartVkOAuth = () => {
    const state = generateVkOAuthState();
    sessionStorage.setItem(VK_OAUTH_STATE_KEY, state);

    try {
      window.location.assign(buildVkImplicitAuthorizeUrl(state));
    } catch {
      /**
       * `buildVkImplicitAuthorizeUrl` throws when env vars are incomplete.
       */
    }
  };

  if (!auth.accessToken) {
    return null;
  }

  const vkUnavailable =
    connectionQuery.isError &&
    isApiError(connectionQuery.error) &&
    connectionQuery.error.status === 503;

  if (vkUnavailable) {
    return (
      <section className="profile-card">
        <div className="profile-card__header">
          <div>
            <p className="eyebrow">VK</p>
            <h2>ВКонтакте</h2>
            <p className="page-copy">
              Интеграция недоступна: на сервере отключены VK-сервисы или сервис отвечает 503.
            </p>
          </div>
        </div>
        <Notice>
          Если вы администратор, включите <code className="inline-code">VK_ENABLED</code> и
          настройте сохранение токенов в бэкенде.
        </Notice>
      </section>
    );
  }

  if (connectionQuery.isPending || connectionQuery.isFetching) {
    return (
      <section className="profile-card vk-panel vk-panel--loading">
        <p className="eyebrow">VK</p>
        <h2>ВКонтакте</h2>
        <p className="page-copy">Загружаем состояние подключения…</p>
      </section>
    );
  }

  if (connectionQuery.isError) {
    const message =
      connectionQuery.error instanceof Error
        ? connectionQuery.error.message
        : 'Неизвестная ошибка';
    const friendly = isApiError(connectionQuery.error)
      ? getUserFacingApiErrorMessage(connectionQuery.error)
      : message;

    return (
      <section className="profile-card vk-panel">
        <div className="profile-card__header">
          <div>
            <p className="eyebrow">VK</p>
            <h2>ВКонтакте</h2>
          </div>
        </div>
        <Notice>{friendly}</Notice>
        <button
          className={buttonClassName({ variant: 'ghost' })}
          disabled={connectionQuery.isFetching}
          onClick={() => void connectionQuery.refetch()}
          type="button"
        >
          Повторить
        </button>
      </section>
    );
  }

  const vk = connectionQuery.data.data.connection;

  const isLinked = vk.state !== 'disconnected';
  const canSync =
    isLinked &&
    vk.consent.granted &&
    vk.credential.configured &&
    vk.state !== 'disconnected';

  return (
    <section className="profile-card vk-panel">
      <div className="profile-card__header">
        <div>
          <p className="eyebrow">VK</p>
          <h2>ВКонтакте</h2>
          <p className="page-copy">
            Свяжите аккаунт, чтобы сохранять токен на сервере и импортировать ваши группы ВКонтакте как интересы.
          </p>
        </div>
      </div>

      <dl className="profile-list vk-panel__stats">
        <div>
          <dt>Статус</dt>
          <dd>{mapConnectionUiState(vk.state)}</dd>
        </div>
        {isLinked && vk.profile.provider_user_id ? (
          <div>
            <dt>ID в VK</dt>
            <dd>{vk.profile.provider_user_id}</dd>
          </div>
        ) : null}
        {isLinked && vk.profile.screen_name ? (
          <div>
            <dt>Ник</dt>
            <dd>{vk.profile.screen_name}</dd>
          </div>
        ) : null}
        {isLinked ? (
          <div>
            <dt>Токен на сервере</dt>
            <dd>{vk.credential.configured ? 'Сохранён' : 'Не передан — синхронизация недоступна'}</dd>
          </div>
        ) : null}
        {isLinked ? (
          <div>
            <dt>Синхронизация</dt>
            <dd>{mapSyncUiStatus(vk.sync.last_status)}</dd>
          </div>
        ) : null}
        {vk.sync.last_synced_at ? (
          <div>
            <dt>Последний импорт</dt>
            <dd>{formatDateTime(vk.sync.last_synced_at)}</dd>
          </div>
        ) : null}
      </dl>

      {vk.sync.last_error_code ? (
        <Notice>
          Последний код ошибки: <code className="inline-code">{vk.sync.last_error_code}</code>
        </Notice>
      ) : null}

      {isLinked && vk.imported_interests.length ? (
        <details className="vk-panel__interests">
          <summary>
            Импортированные интересы ({vk.imported_interests.length})
          </summary>
          <ul>
            {vk.imported_interests.slice(0, 25).map((item) => (
              <li key={`${item.normalized_name}-${item.position}-${item.imported_at}`}>
                <span>{item.name}</span>
                {item.source_label ? (
                  <span className="vk-panel__source">{` · ${item.source_label}`}</span>
                ) : null}
              </li>
            ))}
          </ul>
          {vk.imported_interests.length > 25 ? (
            <p className="page-copy">Показаны первые 25 записей.</p>
          ) : null}
        </details>
      ) : null}

      <div className="vk-panel__actions">
        {canSync ? (
          <Button
            disabled={syncMutation.isPending}
            onClick={() => {
              syncMutation.reset();
              void syncMutation.mutateAsync();
            }}
            type="button"
            variant="secondary"
          >
            {syncMutation.isPending ? 'Синхронизируем…' : 'Синхронизировать интересы'}
          </Button>
        ) : null}

        {isLinked ? (
          <Button
            disabled={disconnectMutation.isPending}
            onClick={() => {
              disconnectMutation.reset();
              void disconnectMutation.mutateAsync();
            }}
            type="button"
            variant="ghost"
          >
            {disconnectMutation.isPending ? 'Отключаем…' : 'Отключить VK'}
          </Button>
        ) : null}
      </div>

      {syncMutation.isError ? (
        <ErrorBanner
          error={syncMutation.error}
          title="Синхронизация не удалась"
        />
      ) : null}
      {disconnectMutation.isError ? (
        <ErrorBanner
          error={disconnectMutation.error}
          title="Не удалось отключить VK"
        />
      ) : null}

      {!isLinked ? (
        isVkOAuthConfigured() ? (
          <div className="vk-panel__oauth">
            {connectMutation.isError ? (
              <ErrorBanner
                error={connectMutation.error}
                title="Не удалось сохранить подключение"
              />
            ) : null}
            {connectMutation.isSuccess ? (
              <Notice tone="success">Подключение VK обновлено.</Notice>
            ) : null}
            <Button
              onClick={handleStartVkOAuth}
              type="button"
            >
              Войти через VK
            </Button>
          </div>
        ) : (
          <form
            className="profile-form vk-panel__form"
            onSubmit={manualForm.handleSubmit((values) => {
              connectMutation.reset();
              void connectMutation.mutateAsync(buildManualConnectBody(values));
            })}
          >
            {connectMutation.isError ? (
              <ErrorBanner
                error={connectMutation.error}
                title="Не удалось сохранить подключение"
              />
            ) : null}
            {connectMutation.isSuccess ? (
              <Notice tone="success">Подключение VK обновлено.</Notice>
            ) : null}

            <Field
              error={manualForm.formState.errors.provider_user_id?.message}
              hint="Числовой id пользователя VK"
              label="Идентификатор VK"
            >
              <Input
                autoComplete="off"
                placeholder="Например, 123456789"
                {...manualForm.register('provider_user_id')}
              />
            </Field>

            <Field
              error={manualForm.formState.errors.access_token?.message}
              hint="Секретный токен — не публикуйте его"
              label="Access token (необязательно)"
            >
              <textarea
                className={cn('input', 'vk-panel__token')}
                placeholder="Вставьте access token"
                rows={3}
                {...manualForm.register('access_token')}
              />
            </Field>

            <Field
              error={manualForm.formState.errors.scopes_csv?.message}
              hint="Например: friends, groups"
              label="Scopes (через запятую)"
            >
              <Input
                placeholder="friends, groups"
                {...manualForm.register('scopes_csv')}
              />
            </Field>

            <Field
              error={manualForm.formState.errors.expires_at_local?.message}
              hint="Локальное время окончания токена"
              label="Истекает"
            >
              <Input
                type="datetime-local"
                {...manualForm.register('expires_at_local')}
              />
            </Field>

            <Field
              error={manualForm.formState.errors.screen_name?.message}
              label="Ник (необязательно)"
            >
              <Input autoComplete="off" {...manualForm.register('screen_name')} />
            </Field>

            <Field
              error={manualForm.formState.errors.profile_url?.message}
              label="Ссылка на профиль (необязательно)"
            >
              <Input
                autoComplete="off"
                placeholder="https://vk.com/id123"
                type="url"
                {...manualForm.register('profile_url')}
              />
            </Field>

            <label className="vk-panel__consent">
              <Controller
                control={manualForm.control}
                name="consent_acknowledged"
                render={({ field }) => (
                  <input
                    checked={field.value}
                    className="vk-panel__consent-input"
                    onBlur={field.onBlur}
                    onChange={(event) => field.onChange(event.currentTarget.checked)}
                    ref={field.ref}
                    type="checkbox"
                  />
                )}
              />
              <span>
                Я подтверждаю согласие на обработку данных VK в рамках этой интеграции
                (версия политики {VK_CONSENT_POLICY_VERSION}).
              </span>
            </label>
            {manualForm.formState.errors.consent_acknowledged?.message ? (
              <span className="field__error">
                {manualForm.formState.errors.consent_acknowledged.message}
              </span>
            ) : null}

            <div className="profile-actions">
              <Button disabled={connectMutation.isPending} type="submit">
                {connectMutation.isPending ? 'Сохраняем…' : 'Сохранить подключение'}
              </Button>
            </div>
          </form>
        )
      ) : null}
    </section>
  );
}
