import { PasswordResetRequestForm } from '../../../features/auth/ui/password-reset-request-form';
import { Container } from '../../../shared/ui/layout/container';

export function PasswordResetPage(): JSX.Element {
  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Password reset</p>
        <h1>Запросите восстановление доступа.</h1>
        <p>
          Текущий backend поддерживает только request endpoint. Подтверждающий
          reset flow на отдельном экране пока не реализован в API.
        </p>
      </section>

      <PasswordResetRequestForm />
    </Container>
  );
}
