import type {
  ListCatalogCategoriesQuery,
  ListCatalogCategoriesResponse,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';
import { filterVisibleCategories } from '../../../shared/lib/category-visibility';

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

  if (query.has_gifts !== undefined) {
    search.set('has_gifts', String(query.has_gifts));
  }

  const suffix = search.size > 0 ? `?${search.toString()}` : '';

  return requestJson<ListCatalogCategoriesResponse>(
    `/api/v1/catalog/categories${suffix}`,
  ).then((response) => ({
    ...response,
    data: {
      ...response.data,
      items: filterVisibleCategories(response.data.items),
    },
  }));
}
