import { defaultThresholds } from './lib/config.js';
import { weightedPick } from './lib/http.js';
import { loadTestData } from './lib/setup.js';
import {
  authenticatedProfileFlow,
  catalogBrowseFlow,
  healthFlow,
  recommendFlow,
  thinkTime,
} from './lib/scenarios.js';

export const options = {
  scenarios: {
    mixed_load: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: __ENV.K6_RAMP_UP || '1m', target: Number(__ENV.K6_VUS || 20) },
        { duration: __ENV.K6_STEADY || '3m', target: Number(__ENV.K6_VUS || 20) },
        { duration: __ENV.K6_RAMP_DOWN || '30s', target: 0 },
      ],
      gracefulRampDown: '30s',
    },
  },
  thresholds: defaultThresholds(),
};

export function setup() {
  return loadTestData();
}

export default function (data) {
  const action = weightedPick([
    { weight: 5, fn: () => healthFlow() },
    { weight: 40, fn: () => catalogBrowseFlow(data) },
    { weight: 50, fn: () => recommendFlow(data) },
    { weight: 5, fn: () => authenticatedProfileFlow(data) },
  ]);
  action();
  thinkTime(0.5, 2);
}
