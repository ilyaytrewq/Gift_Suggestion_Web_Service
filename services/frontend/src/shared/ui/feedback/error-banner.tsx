import { getUserFacingApiErrorMessage } from '../../api/api-error';
import { isApiError } from '../../api/http';

export function ErrorBanner({
  error,
  title = 'Не удалось выполнить действие',
}: {
  error: unknown;
  title?: string;
}): JSX.Element {
  const description = isApiError(error)
    ? getUserFacingApiErrorMessage(error)
    : 'Что-то пошло не так. Попробуйте ещё раз.';

  return (
    <div className="banner banner--error" role="alert">
      <strong>{title}</strong>
      <p>{description}</p>
    </div>
  );
}
