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

const AGE_OPTIONS: Array<{ label: string; value: 0 | 12 | 16 | 18 }> = [
  { label: '0+', value: 0 },
  { label: '12+', value: 12 },
  { label: '16+', value: 16 },
  { label: '18+', value: 18 },
];

export function CatalogFilters({
  categories,
  filters,
  onAgeRestrictionChange,
  onCategoryChange,
  onClear,
  onHasImageChange,
  onMaxPriceChange,
  onMinPriceChange,
  onSearch,
  onSortChange,
}: {
  categories: CatalogCategory[];
  filters: ListCatalogGiftsQuery;
  onAgeRestrictionChange: (value: 0 | 12 | 16 | 18 | null) => void;
  onCategoryChange: (value: string | null) => void;
  onClear: () => void;
  onHasImageChange: (value: boolean | null) => void;
  onMaxPriceChange: (value: string | null) => void;
  onMinPriceChange: (value: string | null) => void;
  onSearch: (value: string) => void;
  onSortChange: (value: NonNullable<ListCatalogGiftsQuery['sort']>) => void;
}): JSX.Element {
  const hasActiveFilters = Boolean(
    filters.q ||
      filters.category_id ||
      filters.min_price ||
      filters.max_price ||
      filters.age_restriction !== undefined ||
      filters.has_image,
  );

  return (
    <section className="catalog-filters">
      <form
        className="catalog-filters__toolbar"
        onSubmit={(event) => {
          event.preventDefault();
          const formData = new FormData(event.currentTarget);
          onSearch(String(formData.get('q') ?? ''));
          onMinPriceChange(String(formData.get('min_price') ?? '').trim() || null);
          onMaxPriceChange(String(formData.get('max_price') ?? '').trim() || null);
        }}
      >
        <Input
          aria-label="Поиск подарков"
          defaultValue={filters.q ?? ''}
          name="q"
          placeholder="Поиск по каталогу"
        />

        <Input
          aria-label="Цена от"
          defaultValue={filters.min_price ?? ''}
          min="0"
          name="min_price"
          placeholder="Цена от"
          style={{ width: '110px' }}
          type="number"
        />

        <Input
          aria-label="Цена до"
          defaultValue={filters.max_price ?? ''}
          min="0"
          name="max_price"
          placeholder="Цена до"
          style={{ width: '110px' }}
          type="number"
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
        <select
          aria-label="Ограничение по возрасту"
          className="select"
          style={{ width: 'auto' }}
          value={filters.age_restriction !== undefined ? String(filters.age_restriction) : ''}
          onChange={(e) => {
            const val = e.target.value;
            onAgeRestrictionChange(val !== '' ? (Number(val) as 0 | 12 | 16 | 18) : null);
          }}
        >
          <option value="">Любой возраст</option>
          {AGE_OPTIONS.map((opt) => (
            <option key={opt.value} value={String(opt.value)}>
              {opt.label}
            </option>
          ))}
        </select>

        <label
          style={{
            alignItems: 'center',
            cursor: 'pointer',
            display: 'flex',
            fontSize: '0.875rem',
            gap: '0.35rem',
          }}
        >
          <input
            checked={filters.has_image ?? false}
            type="checkbox"
            onChange={(e) => onHasImageChange(e.target.checked || null)}
          />
          Только с фото
        </label>

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

        {hasActiveFilters ? (
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
