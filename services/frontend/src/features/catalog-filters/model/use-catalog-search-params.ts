import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

import type { ListCatalogGiftsQuery } from '../../../shared/api/contracts';

const DEFAULT_SORT = 'newest';

export function useCatalogSearchParams(): {
  filters: ListCatalogGiftsQuery;
  setCategoryId: (value: string | null) => void;
  setQuery: (value: string) => void;
  setSort: (value: NonNullable<ListCatalogGiftsQuery['sort']>) => void;
  clearFilters: () => void;
} {
  const [searchParams, setSearchParams] = useSearchParams();

  const filters = useMemo<ListCatalogGiftsQuery>(
    () => ({
      category_id: searchParams.get('category_id') || undefined,
      limit: 12,
      offset: 0,
      q: searchParams.get('q') || undefined,
      sort:
        (searchParams.get('sort') as ListCatalogGiftsQuery['sort'] | null) ??
        DEFAULT_SORT,
    }),
    [searchParams],
  );

  return {
    filters,
    setCategoryId: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value) {
        next.set('category_id', value);
      } else {
        next.delete('category_id');
      }
      setSearchParams(next);
    },
    setQuery: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value.trim()) {
        next.set('q', value.trim());
      } else {
        next.delete('q');
      }
      setSearchParams(next);
    },
    setSort: (value) => {
      const next = new URLSearchParams(searchParams);
      next.set('sort', value);
      setSearchParams(next);
    },
    clearFilters: () => {
      setSearchParams(
        new URLSearchParams({
          sort: DEFAULT_SORT,
        }),
      );
    },
  };
}
