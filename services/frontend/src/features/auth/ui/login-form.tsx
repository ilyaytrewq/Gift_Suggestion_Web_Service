import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import {
  Link,
  useNavigate,
  useSearchParams,
} from 'react-router-dom';

import { applyAuthenticatedSession, loginUser } from '../api/auth';
import { loginSchema, type LoginSchema } from '../model/schemas';
import { useAuth } from '../../../shared/auth/use-auth';
import { Button } from '../../../shared/ui/button/button';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';
import { PasswordInput } from '../../../shared/ui/input/password-input';

function resolvePostLoginPath(nextPath: string | null): string {
  if (!nextPath) {
    return '/';
  }

  if (!nextPath.startsWith('/')) {
    return '/';
  }

  if (nextPath.startsWith('//')) {
    return '/';
  }

  return nextPath;
}

export function LoginForm(): JSX.Element {
  const auth = useAuth();
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const postLoginPath = resolvePostLoginPath(searchParams.get('next'));
  const form = useForm<LoginSchema>({
    defaultValues: {
      email: '',
      password: '',
    },
    resolver: zodResolver(loginSchema),
  });

  const mutation = useMutation({
    mutationFn: loginUser,
    onSuccess: (payload) => {
      applyAuthenticatedSession(payload.data.auth.access_token);
      auth.setSession(payload.data.auth.access_token, payload.data.user);
      navigate(postLoginPath, { replace: true });
    },
  });

  return (
    <form
      className="auth-form"
      onSubmit={form.handleSubmit((values) => {
        mutation.reset();
        void mutation.mutateAsync(values);
      })}
    >
      {mutation.isError ? <ErrorBanner error={mutation.error} /> : null}

      <Field error={form.formState.errors.email?.message} label="Электронная почта">
        <Input
          autoComplete="email"
          placeholder="you@example.com"
          {...form.register('email')}
        />
      </Field>

      <Field
        error={form.formState.errors.password?.message}
        hint="От 8 символов: строчные и заглавные буквы, цифра, спецсимвол (!@#$% и др.), без пробелов"
        label="Пароль"
      >
        <PasswordInput
          autoComplete="current-password"
          placeholder="Введите пароль"
          {...form.register('password')}
        />
      </Field>

      <Button disabled={mutation.isPending} type="submit">
        {mutation.isPending ? 'Входим...' : 'Войти'}
      </Button>

      <div className="auth-form__links">
        <div className="auth-form__register-block">
          <span className="auth-form__register-hint">Нет аккаунта?</span>
          <Link className="auth-form__link-action" to="/register">
            Зарегистрироваться
          </Link>
        </div>
        <Link className="auth-form__link-action" to="/password-reset">
          Забыли пароль?
        </Link>
      </div>
    </form>
  );
}
