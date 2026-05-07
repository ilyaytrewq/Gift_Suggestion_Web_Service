import { useSearchParams } from 'react-router-dom';

import { LoginForm } from '../../../features/auth/ui/login-form';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Container } from '../../../shared/ui/layout/container';

export function LoginPage(): JSX.Element {
  const [searchParams] = useSearchParams();
  const passwordReset = searchParams.get('password_reset') === '1';

  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Вход</p>
        <h1>Войдите в аккаунт.</h1>
        <p>
          Так вы сможете быстрее вернуться к подбору и управлять данными профиля.
        </p>
        {passwordReset && (
          <Notice tone="success">
            Пароль успешно изменён. Войдите с новым паролем.
          </Notice>
        )}
      </section>

      <LoginForm />
    </Container>
  );
}
