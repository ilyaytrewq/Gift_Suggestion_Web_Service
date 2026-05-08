import type {
  ConnectVKRequestBody,
  VKConnectionEnvelope,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';

export function getVkConnection(): Promise<VKConnectionEnvelope> {
  return requestJson<VKConnectionEnvelope>('/api/v1/integrations/vk/connection', {
    auth: true,
  });
}

export function connectVk(
  body: ConnectVKRequestBody,
): Promise<VKConnectionEnvelope> {
  return requestJson<VKConnectionEnvelope>('/api/v1/integrations/vk/connection', {
    method: 'PUT',
    auth: true,
    body,
  });
}

export function disconnectVk(): Promise<VKConnectionEnvelope> {
  return requestJson<VKConnectionEnvelope>('/api/v1/integrations/vk/connection', {
    method: 'DELETE',
    auth: true,
  });
}

export function syncVkInterests(): Promise<VKConnectionEnvelope> {
  return requestJson<VKConnectionEnvelope>(
    '/api/v1/integrations/vk/connection/sync-interests',
    {
      method: 'POST',
      auth: true,
    },
  );
}
