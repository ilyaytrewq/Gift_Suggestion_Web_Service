import type {
  RecommendationRequest,
  RecommendationResponse,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';
import { sanitizeItemCategory } from '../../../shared/lib/category-visibility';

export function createRecommendation(
  payload: RecommendationRequest,
): Promise<RecommendationResponse> {
  return requestJson<RecommendationResponse>('/api/v1/recommendations', {
    method: 'POST',
    auth: true,
    body: payload,
  }).then((response) => ({
    ...response,
    data: {
      ...response.data,
      recommendation: {
        ...response.data.recommendation,
        recommendations: response.data.recommendation.recommendations.map((item) => ({
          ...item,
          gift: sanitizeItemCategory(item.gift),
          alternatives: item.alternatives.map((alternative) => ({
            ...alternative,
            gift: sanitizeItemCategory(alternative.gift),
          })),
        })),
      },
    },
  }));
}
