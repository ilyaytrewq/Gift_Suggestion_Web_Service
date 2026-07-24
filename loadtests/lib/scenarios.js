import http from 'k6/http';
import { sleep } from 'k6';

import { apiPrefix, authHeaders, baseUrl, jsonHeaders } from './config.js';
import { buildRecommendPayload, catalogListQuery } from './payloads.js';
import { checkOk, pickRandom } from './http.js';

export function healthFlow() {
  const live = http.get(`${baseUrl}/health/live`, {
    tags: { endpoint: 'health' },
  });
  checkOk(live, 'health live');

  const ready = http.get(`${baseUrl}/health/ready`, {
    tags: { endpoint: 'health' },
  });
  checkOk(ready, 'health ready');
}

export function catalogBrowseFlow(data) {
  const categories = http.get(`${apiPrefix}/catalog/categories?limit=20`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'catalog_categories' },
  });
  checkOk(categories, 'catalog categories');

  const list = http.get(`${apiPrefix}/catalog/gifts?${catalogListQuery()}`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'catalog_list' },
  });
  checkOk(list, 'catalog gifts list');

  const giftId = pickRandom(data.giftIds);
  if (!giftId) {
    return;
  }

  const gift = http.get(`${apiPrefix}/catalog/gifts/${giftId}`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'catalog_gift' },
  });
  checkOk(gift, 'catalog gift');

  if (Math.random() < 0.4) {
    const similar = http.get(`${apiPrefix}/catalog/gifts/${giftId}/similar?limit=6`, {
      headers: jsonHeaders(),
      tags: { endpoint: 'catalog_similar' },
    });
    checkOk(similar, 'catalog similar');
  }
}

export function recommendFlow(data) {
  const headers = authHeaders(data.accessToken);
  const res = http.post(`${apiPrefix}/recommendations`, buildRecommendPayload(), {
    headers,
    tags: { endpoint: 'recommendations' },
  });
  checkOk(res, 'recommendations create');
}

export function authenticatedProfileFlow(data) {
  if (!data.accessToken) {
    return;
  }

  const me = http.get(`${apiPrefix}/users/me`, {
    headers: authHeaders(data.accessToken),
    tags: { endpoint: 'users_me' },
  });
  checkOk(me, 'users me');
}

export function thinkTime(minSec = 0.3, maxSec = 1.2) {
  sleep(minSec + Math.random() * (maxSec - minSec));
}
