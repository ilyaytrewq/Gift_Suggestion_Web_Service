import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';

import { applyAuthenticatedSession, loginUser } from '../api/auth';
import { loginSchema, type LoginSchema } from '../model/schemas';
import { useAuth } from '../../../shared/auth/use-auth';
import { Button } from '../../../shared/ui/button/button';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';

export function LoginForm(): JSX.Element {
  const auth = useAuth();
  const navigate = useNavigate();
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
      navigate('/');
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

      <Field error={form.formState.errors.email?.message} label="Email">
        <Input
          autoComplete="email"
          placeholder="you@example.com"
          {...form.register('email')}
        />
      </Field>

      <Field error={form.formState.errors.password?.message} label="Пароль">
        <Input
          autoComplete="current-password"
          placeholder="Введите пароль"
          type="password"
          {...form.register('password')}
        />
      </Field>

      <Button disabled={mutation.isPending} type="submit">
        {mutation.isPending ? 'Входим...' : 'Войти'}
      </Button>

      <div className="auth-form__links">
        <Link to="/register">Нет аккаунта? Зарегистрироваться</Link>
        <Link to="/password-reset">Забыли пароль?</Link>
      </div>
    </form>
  );
}
