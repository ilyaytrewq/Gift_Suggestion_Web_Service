import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';

import { listCatalogCategories } from '../../../entities/category/api/categories';
import { listCatalogGifts } from '../../../entities/gift/api/gifts';
import { GiftPreviewCard } from '../../../entities/gift/ui/gift-preview-card';
import {
  CATALOG_PAGE_SIZE,
  useCatalogSearchParams,
} from '../../../features/catalog-filters/model/use-catalog-search-params';
import { CatalogFilters } from '../../../features/catalog-filters/ui/catalog-filters';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { EmptyState } from '../../../shared/ui/feedback/empty-state';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Container } from '../../../shared/ui/layout/container';
import { CatalogGridSkeleton } from '../../../shared/ui/skeleton/skeleton';

function buildPagination(currentPage: number, totalPages: number): Array<number | 'ellipsis'> {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, index) => index + 1);
  }

  let start = Math.max(2, currentPage - 1);
  let end = Math.min(totalPages - 1, currentPage + 1);

  if (currentPage <= 3) {
    end = 4;
  }

  if (currentPage >= totalPages - 2) {
    start = totalPages - 3;
  }

  const items: Array<number | 'ellipsis'> = [1];

  if (start > 2) {
    items.push('ellipsis');
  }

  for (let value = start; value <= end; value += 1) {
    items.push(value);
  }

  if (end < totalPages - 1) {
    items.push('ellipsis');
  }

  items.push(totalPages);

  return items;
}

export function CatalogPage(): JSX.Element {
  const { clearFilters, filters, page, setCategoryId, setPage, setQuery, setSort } =
    useCatalogSearchParams();

  const categoriesQuery = useQuery({
    queryFn: () =>
      listCatalogCategories({
        limit: 100,
        offset: 0,
        sort: 'name_asc',
        has_gifts: true,
      }),
    queryKey: ['categories', 'catalog', 'with-gifts'],
  });

  const giftsQuery = useQuery({
    queryFn: () => listCatalogGifts(filters),
    queryKey: ['catalog', filters],
  });

  const giftItems = giftsQuery.data?.data.items ?? [];
  const pageInfo = giftsQuery.data?.data.page;
  const totalGifts = pageInfo?.total ?? 0;
  const pageSize = pageInfo?.limit ?? filters.limit ?? CATALOG_PAGE_SIZE;
  const totalPages = totalGifts > 0 ? Math.ceil(totalGifts / pageSize) : 0;
  const currentPage = totalPages > 0 ? Math.min(page, totalPages) : 1;
  const shownFrom = totalGifts > 0 && pageInfo ? pageInfo.offset + 1 : 0;
  const shownTo = totalGifts > 0 && pageInfo ? pageInfo.offset + giftItems.length : 0;
  const shownRange = totalGifts > 0 ? `${shownFrom}-${shownTo}` : '0';
  const paginationItems =
    totalPages > 1 ? buildPagination(currentPage, totalPages) : [];

  return (
    <Container className="page-stack">
      <section className="section-heading">
        <p className="eyebrow">Каталог идей</p>
        <h1>Выберите подарок по категории, названию и цене.</h1>
        <p className="page-copy">
          Используйте поиск и фильтры, чтобы быстро сузить выбор.
        </p>
      </section>

      {categoriesQuery.data ? (
        <CatalogFilters
          categories={categoriesQuery.data.data.items}
          filters={filters}
          onCategoryChange={setCategoryId}
          onClear={clearFilters}
          onSearch={setQuery}
          onSortChange={setSort}
        />
      ) : null}

      {categoriesQuery.isError ? (
        <ErrorBanner error={categoriesQuery.error} title="Не удалось загрузить категории" />
      ) : null}

      {giftsQuery.isError ? (
        <ErrorBanner error={giftsQuery.error} title="Не удалось загрузить каталог" />
      ) : null}

      {giftsQuery.isLoading ? <CatalogGridSkeleton /> : null}

      {giftsQuery.data ? (
        <>
          <div className="catalog-summary">
            <span>Подарков найдено: {totalGifts}</span>
            <span>Показаны: {shownRange}</span>
            <span>
              Страница: {currentPage}
              {totalPages > 0 ? ` из ${totalPages}` : ''}
            </span>
          </div>

          {giftItems.length ? (
            <div className="gift-grid">
              {giftItems.map((gift, index) => (
                <GiftPreviewCard
                  gift={gift}
                  key={gift.id}
                  listPosition={
                    pageInfo ? pageInfo.offset + index + 1 : undefined
                  }
                />
              ))}
            </div>
          ) : totalGifts > 0 ? (
            <EmptyState
              action={
                <button
                  className={buttonClassName({ variant: 'secondary' })}
                  onClick={() => {
                    setPage(totalPages);
                  }}
                  type="button"
                >
                  Открыть последнюю страницу
                </button>
              }
              description="По текущим фильтрам товары есть, но на этой странице их уже нет."
              title="Страница каталога пуста"
            />
          ) : (
            <EmptyState
              action={
                <Link
                  className={buttonClassName({ variant: 'secondary' })}
                  to="/"
                >
                  Вернуться на главную
                </Link>
              }
              description="Попробуйте изменить запрос, сбросить категорию или выбрать другой тип сортировки."
              title="Ничего не найдено"
            />
          )}

          {paginationItems.length ? (
            <nav
              aria-label="Навигация по страницам каталога"
              className="catalog-pagination"
            >
              <button
                className={buttonClassName({ variant: 'ghost' })}
                disabled={page <= 1}
                onClick={() => {
                  setPage(page - 1);
                }}
                type="button"
              >
                Назад
              </button>

              <div className="catalog-pagination__pages">
                {paginationItems.map((item, index) =>
                  item === 'ellipsis' ? (
                    <span
                      aria-hidden="true"
                      className="catalog-pagination__ellipsis"
                      key={`ellipsis-${index}`}
                    >
                      …
                    </span>
                  ) : (
                    <button
                      aria-current={item === currentPage ? 'page' : undefined}
                      className={[
                        buttonClassName({
                          variant: item === currentPage ? 'secondary' : 'ghost',
                        }),
                        'catalog-pagination__page',
                      ]
                        .filter(Boolean)
                        .join(' ')}
                      key={item}
                      onClick={() => {
                        setPage(item);
                      }}
                      type="button"
                    >
                      {item}
                    </button>
                  ),
                )}
              </div>

              <button
                className={buttonClassName({ variant: 'ghost' })}
                disabled={page >= totalPages}
                onClick={() => {
                  setPage(page + 1);
                }}
                type="button"
              >
                Вперёд
              </button>
            </nav>
          ) : null}
        </>
      ) : null}
    </Container>
  );
}
