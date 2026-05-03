import { useCallback } from 'react';

import { makeClientEventId, trackEvent } from '../api/tracking';
import type { TrackEventRequest } from '../../../shared/api/contracts';

type TrackParams = Omit<TrackEventRequest, 'client_event_id' | 'occurred_at'>;

export function useTrackEvent() {
  return useCallback((params: TrackParams) => {
    void trackEvent({
      ...params,
      client_event_id: makeClientEventId(),
      occurred_at: new Date().toISOString(),
    }).catch(() => {
      // tracking failures are silent
    });
  }, []);
}
