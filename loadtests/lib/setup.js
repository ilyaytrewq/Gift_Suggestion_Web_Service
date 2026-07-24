import http from 'k6/http';

import { apiPrefix, baseUrl, jsonHeaders, testEmail, testPassword, accessToken } from './config.js';
import { extractGiftIds, parseJson } from './http.js';

export function loadTestData() {
  const giftsRes = http.get(`${apiPrefix}/catalog/gifts?limit=24&has_image=true`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'catalog_list', phase: 'setup' },
  });

  const giftsBody = parseJson(giftsRes);
  const giftIds = extractGiftIds(giftsBody);

  let token = accessToken;
  if (!token && testEmail && testPassword) {
    const loginRes = http.post(
      `${apiPrefix}/auth/login`,
      JSON.stringify({ email: testEmail, password: testPassword }),
      { headers: jsonHeaders(), tags: { endpoint: 'auth_login', phase: 'setup' } },
    );
    const loginBody = parseJson(loginRes);
    token = loginBody?.data?.access_token || '';
  }

  return {
    baseUrl,
    giftIds,
    accessToken: token,
    catalogEmpty: giftIds.length === 0,
  };
}
