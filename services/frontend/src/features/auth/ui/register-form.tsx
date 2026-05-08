import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';

import { registerUser } from '../api/auth';
import { registerSchema, type RegisterSchema } from '../model/schemas';
import { Button } from '../../../shared/ui/button/button';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';
import { PasswordInput } from '../../../shared/ui/input/password-input';

export function RegisterForm(): JSX.Element {
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
  });

  const registeredEmail = form.getValues('email');

  if (mutation.isSuccess) {
    return (
      <div className="auth-form">
        <Notice tone="success">
          <strong>Аккаунт создан!</strong>
          <br />
          Мы отправили письмо на <strong>{registeredEmail}</strong> — перейдите по
          ссылке в письме или введите код из письма на странице подтверждения, затем
          войдите в аккаунт.
        </Notice>
        <p style={{ margin: 0, color: 'var(--ink-muted)', fontSize: '0.92rem' }}>
          Не получили письмо? Проверьте папку «Спам».
        </p>
        <div className="auth-form__links">
          <Link className="auth-form__link-action" to="/auth/email-verify">
            Ввести код из письма
          </Link>
          <Link className="auth-form__link-action" to="/login">
            Войти после подтверждения
          </Link>
        </div>
      </div>
    );
  }

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

      <Field
        error={form.formState.errors.password?.message}
        hint="От 8 символов: строчные и заглавные буквы, цифра, спецсимвол (!@#$% и др.), без пробелов"
        label="Пароль"
      >
        <PasswordInput
          autoComplete="new-password"
          placeholder="Например: Secret1!"
          {...form.register('password')}
        />
      </Field>

      <Button disabled={mutation.isPending} type="submit">
        {mutation.isPending ? 'Создаём аккаунт...' : 'Создать аккаунт'}
      </Button>

      <div className="auth-form__links">
        <div className="auth-form__register-block">
          <span className="auth-form__register-hint">Уже есть аккаунт?</span>
          <Link className="auth-form__link-action" to="/login">
            Войти
          </Link>
        </div>
      </div>
    </form>
  );
}
