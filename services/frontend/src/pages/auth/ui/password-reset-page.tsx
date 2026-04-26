import { PasswordResetRequestForm } from '../../../features/auth/ui/password-reset-request-form';
import { Container } from '../../../shared/ui/layout/container';

export function PasswordResetPage(): JSX.Element {
  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Восстановление пароля</p>
        <h1>Восстановите доступ к аккаунту.</h1>
        <p>
          Введите email, и мы отправим инструкции, если аккаунт существует.
        </p>
      </section>

      <PasswordResetRequestForm />
    </Container>
  );
}
