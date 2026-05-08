import { type FormEvent, useEffect, useState } from 'react';
import { Link, useSearchParams } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';

import { confirmEmailVerification } from '../../../features/auth/api/auth';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Container } from '../../../shared/ui/layout/container';
import { Button } from '../../../shared/ui/button/button';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';

export function EmailVerifyPage(): JSX.Element {
  const [searchParams] = useSearchParams();
  const urlToken = searchParams.get('token');
  const [manualToken, setManualToken] = useState('');

  const mutation = useMutation({
    mutationFn: confirmEmailVerification,
  });

  useEffect(() => {
    if (urlToken?.trim() && mutation.isIdle) {
      void mutation.mutate({ token: urlToken.trim() });
    }
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [urlToken]);

  const onManualSubmit = (event: FormEvent<HTMLFormElement>): void => {
    event.preventDefault();
    const token = manualToken.trim();
    if (!token) {
      return;
    }
    mutation.reset();
    void mutation.mutate({ token });
  };

  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Подтверждение email</p>
        <h1>Подтвердите адрес почты</h1>
      </section>

      {!urlToken?.trim() ? (
        <form className="auth-form" onSubmit={onManualSubmit}>
          <p>
            Вставьте код из письма (тот же, что в ссылке) или откройте ссылку из письма
            в браузере.
          </p>
          <Field label="Код подтверждения">
            <Input
              autoComplete="one-time-code"
              name="verification-token"
              onChange={(e) => setManualToken(e.target.value)}
              placeholder="Вставьте код из письма"
              value={manualToken}
            />
          </Field>
          <Button disabled={mutation.isPending || !manualToken.trim()} type="submit">
            {mutation.isPending ? 'Проверяем…' : 'Подтвердить'}
          </Button>
        </form>
      ) : null}

      {urlToken?.trim() && mutation.isPending ? (
        <p>Подтверждаем адрес электронной почты…</p>
      ) : null}

      {mutation.isError ? <ErrorBanner error={mutation.error} /> : null}

      {mutation.isSuccess ? (
        <Notice tone="success">
          Email подтверждён. Теперь можно{' '}
          <Link to="/login">войти</Link>.
        </Notice>
      ) : null}
    </Container>
  );
}
