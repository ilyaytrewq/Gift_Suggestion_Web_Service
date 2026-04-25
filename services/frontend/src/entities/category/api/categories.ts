import type {
  ListCatalogCategoriesQuery,
  ListCatalogCategoriesResponse,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';

export function listCatalogCategories(
  query: ListCatalogCategoriesQuery = {},
): Promise<ListCatalogCategoriesResponse> {
  const search = new URLSearchParams();

  if (query.q) {
    search.set('q', query.q);
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

  return requestJson<ListCatalogCategoriesResponse>(
    `/api/v1/catalog/categories${suffix}`,
  );
}
