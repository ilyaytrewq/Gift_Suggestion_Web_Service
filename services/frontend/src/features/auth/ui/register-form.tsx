import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';

import { registerUser } from '../api/auth';
import { registerSchema, type RegisterSchema } from '../model/schemas';
import { Button } from '../../../shared/ui/button/button';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';

export function RegisterForm(): JSX.Element {
  const navigate = useNavigate();
  const form = useForm<RegisterSchema>({
    defaultValues: {
      display_name: '',
      email: '',
      password: '',
    },
    resolver: zodResolver(registerSchema),
  });

  const mutation = useMutation({
    mutationFn: registerUser,
    onSuccess: () => {
      navigate('/login?registered=1');
    },
  });

  return (
    <form
      className="auth-form"
      onSubmit={form.handleSubmit((values) => {
        mutation.reset();
        void mutation.mutateAsync({
          ...values,
          display_name: values.display_name || undefined,
        });
      })}
    >
      {mutation.isError ? <ErrorBanner error={mutation.error} /> : null}
      {mutation.isSuccess ? (
        <Notice tone="success">
          Аккаунт создан. Перенаправляем на страницу входа.
        </Notice>
      ) : null}

      <Field
        error={form.formState.errors.display_name?.message}
        hint="Необязательное поле"
        label="Имя"
      >
        <Input
          autoComplete="name"
          placeholder="Например, Илья"
          {...form.register('display_name')}
        />
      </Field>

      <Field error={form.formState.errors.email?.message} label="Электронная почта">
        <Input
          autoComplete="email"
          placeholder="you@example.com"
          {...form.register('email')}
        />
      </Field>

      <Field error={form.formState.errors.password?.message} label="Пароль">
        <Input
          autoComplete="new-password"
          placeholder="Минимум 8 символов"
          type="password"
          {...form.register('password')}
        />
      </Field>

      <Button disabled={mutation.isPending} type="submit">
        {mutation.isPending ? 'Создаём аккаунт...' : 'Создать аккаунт'}
      </Button>

      <div className="auth-form__links">
        <Link to="/login">Уже есть аккаунт? Войти</Link>
      </div>
    </form>
  );
}
