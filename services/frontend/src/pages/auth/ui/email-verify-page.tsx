import { useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';

import { confirmEmailVerification } from '../../../features/auth/api/auth';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Container } from '../../../shared/ui/layout/container';

export function EmailVerifyPage(): JSX.Element {
  const [searchParams] = useSearchParams();
  const token = searchParams.get('token');

  const mutation = useMutation({
    mutationFn: confirmEmailVerification,
  });

  useEffect(() => {
    if (token && mutation.isIdle) {
      void mutation.mutate({ token });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [token]);

  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Подтверждение email</p>
        <h1>Верификация почты.</h1>
      </section>

      {!token && (
        <p>
          Ссылка недействительна.{' '}
          <Link to="/">Вернуться на главную</Link>.
        </p>
      )}

      {mutation.isPending && <p>Подтверждаем адрес электронной почты…</p>}

      {mutation.isError && <ErrorBanner error={mutation.error} />}

      {mutation.isSuccess && (
        <Notice tone="success">
          Email подтверждён. Теперь можно{' '}
          <Link to="/login">войти</Link>.
        </Notice>
      )}
    </Container>
  );
}
