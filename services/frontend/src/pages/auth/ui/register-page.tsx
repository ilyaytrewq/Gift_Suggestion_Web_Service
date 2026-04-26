import { RegisterForm } from '../../../features/auth/ui/register-form';
import { Container } from '../../../shared/ui/layout/container';

export function RegisterPage(): JSX.Element {
  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Регистрация</p>
        <h1>Создайте аккаунт.</h1>
        <p>
          С аккаунтом проще вернуться к подбору и сохранить свои данные.
        </p>
      </section>

      <RegisterForm />
    </Container>
  );
}
