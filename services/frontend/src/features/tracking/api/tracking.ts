import type { TrackEventRequest, TrackEventResponse } from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';

export function trackEvent(payload: TrackEventRequest): Promise<TrackEventResponse> {
  return requestJson<TrackEventResponse>('/api/v1/tracking/events', {
    method: 'POST',
    auth: true,
    body: payload,
  });
}

function randomId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

export function makeClientEventId(): string {
  return randomId();
}
