import {
  useMutation,
  useQuery,
  useQueryClient,
} from '@tanstack/react-query';

import { disconnectVk, getVkConnection, syncVkInterests } from '../api/vk';
import {
  buildVkIdAuthorizeUrl,
  computeVkOAuthCodeChallenge,
  generateVkOAuthCodeVerifier,
  generateVkOAuthState,
  VK_OAUTH_CODE_VERIFIER_KEY,
  VK_OAUTH_STATE_KEY,
} from '../lib/vk-oauth';
import { isVkIntegrationConfigured } from '../lib/is-vk-integration-configured';
import { useAuth } from '../../../shared/auth/use-auth';
import { getUserFacingApiErrorMessage } from '../../../shared/api/api-error';
import { isApiError } from '../../../shared/api/http';
import type { components } from '../../../shared/api/generated/schema';
import { isVkOAuthConfigured } from '../../../shared/config/env';
import { formatDateTime } from '../../../shared/lib/format';
import { Button } from '../../../shared/ui/button/button';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';

type VKConnection = components['schemas']['VKConnection'];

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

export function VkConnectionPanel(): JSX.Element | null {
  const auth = useAuth();
  const queryClient = useQueryClient();
  const oauthConfigured = isVkOAuthConfigured();

  const connectionQuery = useQuery({
    queryKey: ['vk-connection'],
    queryFn: getVkConnection,
    enabled: Boolean(auth.accessToken) && oauthConfigured,
  });

  const disconnectMutation = useMutation({
    mutationFn: disconnectVk,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
    },
  });

  const syncMutation = useMutation({
    mutationFn: syncVkInterests,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ['vk-connection'] });
    },
  });

  const handleStartVkOAuth = async () => {
    const state = generateVkOAuthState();
    const codeVerifier = generateVkOAuthCodeVerifier();
    sessionStorage.setItem(VK_OAUTH_STATE_KEY, state);
    sessionStorage.setItem(VK_OAUTH_CODE_VERIFIER_KEY, codeVerifier);

    try {
      const codeChallenge = await computeVkOAuthCodeChallenge(codeVerifier);
      window.location.assign(buildVkIdAuthorizeUrl(state, codeChallenge));
    } catch {
      /**
       * `buildVkIdAuthorizeUrl` throws when env vars are incomplete.
       */
    }
  };

  if (!auth.accessToken || !oauthConfigured) {
    return null;
  }

  if (connectionQuery.isPending || connectionQuery.isFetching) {
    return null;
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

  if (!isVkIntegrationConfigured(vk)) {
    return null;
  }

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
            Свяжите аккаунт VK ID. Синхронизация подтягивает профиль; список групп появится после выдачи права groups в id.vk.com.
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
        <>
          <ErrorBanner
            error={syncMutation.error}
            title="Синхронизация не удалась"
          />
          {isApiError(syncMutation.error) &&
          (syncMutation.error.code === 'vk_token_expired' ||
            syncMutation.error.code === 'vk_token_invalid') ? (
            <Notice>
              Нажмите «Отключить VK», затем «Войти через VK», чтобы выдать новый токен.
            </Notice>
          ) : null}
        </>
      ) : null}
      {disconnectMutation.isError ? (
        <ErrorBanner
          error={disconnectMutation.error}
          title="Не удалось отключить VK"
        />
      ) : null}

      {!isLinked ? (
        <div className="vk-panel__oauth">
          <Button onClick={handleStartVkOAuth} type="button">
            Войти через VK
          </Button>
        </div>
      ) : null}
    </section>
  );
}
