import { useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';

import type { ListCatalogGiftsQuery } from '../../../shared/api/contracts';

const DEFAULT_SORT = 'newest';
export const CATALOG_PAGE_SIZE = 12;

function parsePage(value: string | null): number {
  if (!value) {
    return 1;
  }

  const page = Number(value);
  if (!Number.isInteger(page) || page < 1) {
    return 1;
  }

  return page;
}

function setPageParam(searchParams: URLSearchParams, value: number): void {
  if (value <= 1) {
    searchParams.delete('page');
    return;
  }

  searchParams.set('page', String(value));
}

export function useCatalogSearchParams(): {
  filters: ListCatalogGiftsQuery;
  page: number;
  setCategoryId: (value: string | null) => void;
  setPage: (value: number) => void;
  setQuery: (value: string) => void;
  setSort: (value: NonNullable<ListCatalogGiftsQuery['sort']>) => void;
  setMinPrice: (value: string | null) => void;
  setMaxPrice: (value: string | null) => void;
  setAgeRestriction: (value: 0 | 12 | 16 | 18 | null) => void;
  setHasImage: (value: boolean | null) => void;
  clearFilters: () => void;
} {
  const [searchParams, setSearchParams] = useSearchParams();
  const page = useMemo(() => parsePage(searchParams.get('page')), [searchParams]);

  const filters = useMemo<ListCatalogGiftsQuery>(
    () => ({
      category_id: searchParams.get('category_id') || undefined,
      limit: CATALOG_PAGE_SIZE,
      offset: (page - 1) * CATALOG_PAGE_SIZE,
      q: searchParams.get('q') || undefined,
      sort:
        (searchParams.get('sort') as ListCatalogGiftsQuery['sort'] | null) ??
        DEFAULT_SORT,
      min_price: searchParams.get('min_price') || undefined,
      max_price: searchParams.get('max_price') || undefined,
      age_restriction: searchParams.get('age_restriction')
        ? (Number(searchParams.get('age_restriction')) as 0 | 12 | 16 | 18)
        : undefined,
      has_image: searchParams.get('has_image') === 'true' ? true : undefined,
    }),
    [page, searchParams],
  );

  return {
    filters,
    page,
    setCategoryId: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value) {
        next.set('category_id', value);
      } else {
        next.delete('category_id');
      }
      setPageParam(next, 1);
      setSearchParams(next);
    },
    setPage: (value) => {
      const next = new URLSearchParams(searchParams);
      setPageParam(next, Math.max(1, Math.trunc(value)));
      setSearchParams(next);
    },
    setQuery: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value.trim()) {
        next.set('q', value.trim());
      } else {
        next.delete('q');
      }
      setPageParam(next, 1);
      setSearchParams(next);
    },
    setSort: (value) => {
      const next = new URLSearchParams(searchParams);
      next.set('sort', value);
      setPageParam(next, 1);
      setSearchParams(next);
    },
    setMinPrice: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value?.trim()) {
        next.set('min_price', value.trim());
      } else {
        next.delete('min_price');
      }
      setPageParam(next, 1);
      setSearchParams(next);
    },
    setMaxPrice: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value?.trim()) {
        next.set('max_price', value.trim());
      } else {
        next.delete('max_price');
      }
      setPageParam(next, 1);
      setSearchParams(next);
    },
    setAgeRestriction: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value !== null) {
        next.set('age_restriction', String(value));
      } else {
        next.delete('age_restriction');
      }
      setPageParam(next, 1);
      setSearchParams(next);
    },
    setHasImage: (value) => {
      const next = new URLSearchParams(searchParams);
      if (value) {
        next.set('has_image', 'true');
      } else {
        next.delete('has_image');
      }
      setPageParam(next, 1);
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
