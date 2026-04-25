import { isApiError } from '../../api/http';

export function ErrorBanner({
  error,
  title = 'Не удалось выполнить запрос',
}: {
  error: unknown;
  title?: string;
}): JSX.Element {
  const description = isApiError(error)
    ? error.message
    : 'Произошла непредвиденная ошибка. Попробуйте ещё раз.';

  return (
    <div className="banner banner--error" role="alert">
      <strong>{title}</strong>
      <p>{description}</p>
    </div>
  );
}
