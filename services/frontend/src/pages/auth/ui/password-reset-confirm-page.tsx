import { useSearchParams } from 'react-router-dom';

import { PasswordResetConfirmForm } from '../../../features/auth/ui/password-reset-confirm-form';
import { Container } from '../../../shared/ui/layout/container';

export function PasswordResetConfirmPage(): JSX.Element {
  const [searchParams] = useSearchParams();
  const urlToken = searchParams.get('token');

  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Восстановление пароля</p>
        <h1>Установите новый пароль.</h1>
        <p>Придумайте надёжный пароль для вашего аккаунта.</p>
      </section>

      <PasswordResetConfirmForm urlToken={urlToken} />
    </Container>
  );
}
