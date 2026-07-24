import http from 'k6/http';
import { check } from 'k6';

import { apiPrefix, baseUrl, jsonHeaders } from './lib/config.js';
import { buildRecommendPayload } from './lib/payloads.js';

const TARGET = __ENV.K6_PROBE_TARGET || 'catalog';
const MAX_RATE = Number(__ENV.K6_MAX_RATE || 30);
const STEP = Number(__ENV.K6_RATE_STEP || 5);
const STEP_SEC = __ENV.K6_STEP_DURATION || '20s';

function buildStages() {
  const stages = [];
  for (let rate = STEP; rate <= MAX_RATE; rate += STEP) {
    stages.push({ duration: STEP_SEC, target: rate });
  }
  stages.push({ duration: '10s', target: 0 });
  return stages;
}

export const options = {
  scenarios: {
    rps_probe: {
      executor: 'ramping-arrival-rate',
      startRate: 1,
      timeUnit: '1s',
      preAllocatedVUs: Number(__ENV.K6_PREALLOCATED_VUS || 80),
      maxVUs: Number(__ENV.K6_MAX_VUS || 150),
      stages: buildStages(),
    },
  },
  thresholds: {
    http_req_failed: ['rate<0.05'],
  },
};

export function setup() {
  if (TARGET === 'health') {
    return {};
  }
  const res = http.get(`${apiPrefix}/catalog/gifts?limit=5`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'setup' },
  });
  if (res.status !== 200) {
    console.warn(`setup catalog list returned ${res.status}`);
  }
  return {};
}

export default function () {
  if (TARGET === 'recommendations') {
    const res = http.post(`${apiPrefix}/recommendations`, buildRecommendPayload({ top_n: 12 }), {
      headers: jsonHeaders(),
      tags: { endpoint: 'recommendations' },
    });
    check(res, { ok: (r) => r.status >= 200 && r.status < 300 });
    return;
  }

  if (TARGET === 'health') {
    const live = http.get(`${baseUrl}/health/live`, { tags: { endpoint: 'health' } });
    check(live, { ok: (r) => r.status === 200 });
    return;
  }

  const res = http.get(`${apiPrefix}/catalog/gifts?limit=12&has_image=true`, {
    headers: jsonHeaders(),
    tags: { endpoint: 'catalog_list' },
  });
  check(res, { ok: (r) => r.status >= 200 && r.status < 300 });
}
