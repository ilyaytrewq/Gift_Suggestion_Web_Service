import type {
  RecommendationRequest,
  RecommendationResponse,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';

export function createRecommendation(
  payload: RecommendationRequest,
): Promise<RecommendationResponse> {
  return requestJson<RecommendationResponse>('/api/v1/recommendations', {
    method: 'POST',
    body: payload,
  });
}
