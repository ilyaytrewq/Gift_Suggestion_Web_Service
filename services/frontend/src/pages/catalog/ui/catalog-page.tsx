import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';

import { listCatalogCategories } from '../../../entities/category/api/categories';
import { listCatalogGifts } from '../../../entities/gift/api/gifts';
import { GiftPreviewCard } from '../../../entities/gift/ui/gift-preview-card';
import { CatalogFilters } from '../../../features/catalog-filters/ui/catalog-filters';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { EmptyState } from '../../../shared/ui/feedback/empty-state';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Container } from '../../../shared/ui/layout/container';
import { CatalogGridSkeleton } from '../../../shared/ui/skeleton/skeleton';
import { useCatalogSearchParams } from '../../../features/catalog-filters/model/use-catalog-search-params';

export function CatalogPage(): JSX.Element {
  const { clearFilters, filters, setCategoryId, setQuery, setSort } =
    useCatalogSearchParams();

  const categoriesQuery = useQuery({
    queryFn: () =>
      listCatalogCategories({
        limit: 20,
        offset: 0,
        sort: 'name_asc',
      }),
    queryKey: ['categories', 'catalog'],
  });

  const giftsQuery = useQuery({
    queryFn: () => listCatalogGifts(filters),
    queryKey: ['catalog', filters],
  });

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
            <span>Подарков найдено: {giftsQuery.data.data.page.total}</span>
            <span>Показано: {giftsQuery.data.data.items.length}</span>
          </div>

          {giftsQuery.data.data.items.length ? (
            <div className="gift-grid">
              {giftsQuery.data.data.items.map((gift) => (
                <GiftPreviewCard gift={gift} key={gift.id} />
              ))}
            </div>
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
        </>
      ) : null}
    </Container>
  );
}
