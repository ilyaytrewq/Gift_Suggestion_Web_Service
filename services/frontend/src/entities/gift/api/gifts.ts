import type {
  GetCatalogGiftResponse,
  GetSimilarGiftsResponse,
  ListCatalogGiftsQuery,
  ListCatalogGiftsResponse,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';
import {
  sanitizeItemCategories,
  sanitizeItemCategory,
} from '../../../shared/lib/category-visibility';

export function listCatalogGifts(
  query: ListCatalogGiftsQuery = {},
): Promise<ListCatalogGiftsResponse> {
  const search = new URLSearchParams();

  if (query.q) {
    search.set('q', query.q);
  }

  if (query.category_id) {
    search.set('category_id', query.category_id);
  }

  if (query.min_price) {
    search.set('min_price', query.min_price);
  }

  if (query.max_price) {
    search.set('max_price', query.max_price);
  }

  if (query.age_restriction !== undefined) {
    search.set('age_restriction', String(query.age_restriction));
  }

  if (query.has_image !== undefined) {
    search.set('has_image', String(query.has_image));
  }

  if (query.limit !== undefined) {
    search.set('limit', String(query.limit));
  }

  if (query.offset !== undefined) {
    search.set('offset', String(query.offset));
  }

  if (query.sort) {
    search.set('sort', query.sort);
  }

  const suffix = search.size > 0 ? `?${search.toString()}` : '';

  return requestJson<ListCatalogGiftsResponse>(`/api/v1/catalog/gifts${suffix}`).then(
    (response) => ({
      ...response,
      data: {
        ...response.data,
        items: sanitizeItemCategories(response.data.items),
      },
    }),
  );
}

export function getCatalogGift(giftId: string): Promise<GetCatalogGiftResponse> {
  return requestJson<GetCatalogGiftResponse>(`/api/v1/catalog/gifts/${giftId}`).then(
    (response) => ({
      ...response,
      data: {
        ...response.data,
        gift: sanitizeItemCategory(response.data.gift),
      },
    }),
  );
}

export function getSimilarGifts(giftId: string, limit = 6): Promise<GetSimilarGiftsResponse> {
  return requestJson<GetSimilarGiftsResponse>(
    `/api/v1/catalog/gifts/${giftId}/similar?limit=${limit}`,
  ).then((response) => ({
    ...response,
    data: {
      ...response.data,
      items: sanitizeItemCategories(response.data.items),
    },
  }));
}
