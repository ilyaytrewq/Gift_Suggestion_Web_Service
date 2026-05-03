type CategoryLike = {
  name: string;
};

const HIDDEN_CATEGORY_NAMES = new Set(['еда и напитки']);

function normalizeCategoryName(name: string): string {
  return name.trim().toLowerCase();
}

export function isVisibleCategoryName(name: string): boolean {
  return !HIDDEN_CATEGORY_NAMES.has(normalizeCategoryName(name));
}

export function filterVisibleCategories<T extends CategoryLike>(categories: T[]): T[] {
  return categories.filter((category) => isVisibleCategoryName(category.name));
}

export function getVisibleCategory<T extends CategoryLike | undefined>(
  category: T,
): T | undefined {
  if (!category) {
    return undefined;
  }

  return isVisibleCategoryName(category.name) ? category : undefined;
}

export function sanitizeItemCategory<T extends { category?: CategoryLike }>(item: T): T {
  const category = getVisibleCategory(item.category);

  if (category === item.category) {
    return item;
  }

  return {
    ...item,
    category,
  } as T;
}

export function sanitizeItemCategories<T extends { category?: CategoryLike }>(items: T[]): T[] {
  return items.map(sanitizeItemCategory);
}
