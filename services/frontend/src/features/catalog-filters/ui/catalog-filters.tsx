import type { CatalogCategory, ListCatalogGiftsQuery } from '../../../shared/api/contracts';
import { Button } from '../../../shared/ui/button/button';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { Input } from '../../../shared/ui/input/input';

const sortOptions: Array<{
  label: string;
  value: NonNullable<ListCatalogGiftsQuery['sort']>;
}> = [
  { label: 'Сначала новые', value: 'newest' },
  { label: 'Название: А-Я', value: 'name_asc' },
  { label: 'Название: Я-А', value: 'name_desc' },
  { label: 'Цена: по возрастанию', value: 'price_asc' },
  { label: 'Цена: по убыванию', value: 'price_desc' },
];

export function CatalogFilters({
  categories,
  filters,
  onCategoryChange,
  onClear,
  onSearch,
  onSortChange,
}: {
  categories: CatalogCategory[];
  filters: ListCatalogGiftsQuery;
  onCategoryChange: (value: string | null) => void;
  onClear: () => void;
  onSearch: (value: string) => void;
  onSortChange: (value: NonNullable<ListCatalogGiftsQuery['sort']>) => void;
}): JSX.Element {
  return (
    <section className="catalog-filters">
      <form
        className="catalog-filters__toolbar"
        onSubmit={(event) => {
          event.preventDefault();
          const formData = new FormData(event.currentTarget);
          onSearch(String(formData.get('q') ?? ''));
        }}
      >
        <Input
          aria-label="Поиск подарков"
          defaultValue={filters.q ?? ''}
          name="q"
          placeholder="Поиск по каталогу"
        />

        <select
          className="select"
          value={filters.sort ?? 'newest'}
          onChange={(event) => {
            onSortChange(
              event.target.value as NonNullable<ListCatalogGiftsQuery['sort']>,
            );
          }}
        >
          {sortOptions.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>

        <Button type="submit">Найти</Button>
      </form>

      <div className="catalog-filters__chips">
        <button
          className={buttonClassName({
            variant: filters.category_id ? 'ghost' : 'secondary',
          })}
          onClick={() => {
            onCategoryChange(null);
          }}
          type="button"
        >
          Все категории
        </button>
        {categories.map((category) => {
          const isAngelina = category.name === 'Специально для Ангелины';
          const isActive = filters.category_id === category.id;
          return (
            <button
              className={[
                buttonClassName({ variant: isActive ? 'secondary' : 'ghost' }),
                isAngelina ? 'button--angelina' : '',
              ]
                .filter(Boolean)
                .join(' ')}
              key={category.id}
              onClick={() => {
                onCategoryChange(category.id);
              }}
              type="button"
            >
              {isAngelina ? '♡ ' : ''}{category.name}
            </button>
          );
        })}

        {(filters.q || filters.category_id) ? (
          <button
            className="catalog-filters__clear"
            onClick={onClear}
            type="button"
          >
            Сбросить фильтры
          </button>
        ) : null}
      </div>
    </section>
  );
}
