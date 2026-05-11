import type { components } from '../../../shared/api/generated/schema';
import { isVkOAuthConfigured } from '../../../shared/config/env';

type VKConnection = components['schemas']['VKConnection'];

export function isVkIntegrationConfigured(connection: VKConnection): boolean {
  return (
    isVkOAuthConfigured() &&
    connection.feature_enabled &&
    connection.token_storage_configured
  );
}
