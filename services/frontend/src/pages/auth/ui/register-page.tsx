import { RegisterForm } from '../../../features/auth/ui/register-form';
import { Container } from '../../../shared/ui/layout/container';

export function RegisterPage(): JSX.Element {
  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Register</p>
        <h1>Создайте аккаунт перед персональными сценариями.</h1>
        <p>
          Профиль, wishlist и recommendation flow будут наращиваться поверх
          этого foundation в следующих срезах.
        </p>
      </section>

      <RegisterForm />
    </Container>
  );
}
