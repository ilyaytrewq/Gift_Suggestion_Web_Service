import { smokeThresholds } from './lib/config.js';
import { weightedPick } from './lib/http.js';
import { loadTestData } from './lib/setup.js';
import {
  catalogBrowseFlow,
  healthFlow,
  recommendFlow,
  thinkTime,
} from './lib/scenarios.js';

export const options = {
  scenarios: {
    smoke: {
      executor: 'constant-vus',
      vus: Number(__ENV.K6_VUS || 3),
      duration: __ENV.K6_DURATION || '30s',
    },
  },
  thresholds: smokeThresholds(),
};

export function setup() {
  return loadTestData();
}

export default function (data) {
  const action = weightedPick([
    { weight: 15, fn: () => healthFlow() },
    { weight: 45, fn: () => catalogBrowseFlow(data) },
    { weight: 40, fn: () => recommendFlow(data) },
  ]);
  action();
  thinkTime(0.2, 0.8);
}
