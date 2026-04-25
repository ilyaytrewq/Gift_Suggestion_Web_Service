import { useSearchParams } from 'react-router-dom';

import { LoginForm } from '../../../features/auth/ui/login-form';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Container } from '../../../shared/ui/layout/container';

export function LoginPage(): JSX.Element {
  const [searchParams] = useSearchParams();
  const wasRegistered = searchParams.get('registered') === '1';

  return (
    <Container className="auth-page">
      <section className="auth-page__intro">
        <p className="eyebrow">Login</p>
        <h1>Войти и продолжить подбор.</h1>
        <p>
          Slice 1 поднимает foundation для auth: login, register, password reset
          request и bootstrap через refresh cookie.
        </p>
        {wasRegistered ? (
          <Notice tone="success">
            Аккаунт создан. Теперь можно войти с теми же реквизитами.
          </Notice>
        ) : null}
      </section>

      <LoginForm />
    </Container>
  );
}
