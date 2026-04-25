const currencyFormatter = new Intl.NumberFormat('ru-RU', {
  style: 'currency',
  currency: 'RUB',
  maximumFractionDigits: 2,
});

export function formatPrice(value: string): string {
  const amount = Number(value);
  if (Number.isNaN(amount)) {
    return value;
  }

  return currencyFormatter.format(amount);
}
