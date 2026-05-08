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

  const metaParts: string[] = [];
  if (isApiError(error) && error.code) {
    metaParts.push(`код ${error.code}`);
  }
  if (isApiError(error) && error.requestId) {
    metaParts.push(`запрос ${error.requestId}`);
  }

  return (
    <div className="banner banner--error" role="alert">
      <strong>{title}</strong>
      <p>{description}</p>
      {metaParts.length > 0 ? (
        <p className="field__hint" style={{ marginTop: '0.35rem' }}>
          {metaParts.join(' · ')}
        </p>
      ) : null}
    </div>
  );
}
