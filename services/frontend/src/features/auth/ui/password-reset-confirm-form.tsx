import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { Link, useNavigate } from 'react-router-dom';

import { confirmPasswordReset } from '../api/auth';
import {
  passwordResetConfirmSchema,
  type PasswordResetConfirmSchema,
} from '../model/schemas';
import { Button } from '../../../shared/ui/button/button';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';
import { PasswordInput } from '../../../shared/ui/input/password-input';

interface Props {
  /** From `?token=`; if missing, user pastes the same code as in the email / link */
  urlToken: string | null;
}

export function PasswordResetConfirmForm({ urlToken }: Props): JSX.Element {
  const trimmedUrlToken = urlToken?.trim() ?? '';
  const [manualToken, setManualToken] = useState('');
  const token = trimmedUrlToken || manualToken.trim();
  const showTokenField = !trimmedUrlToken;

  const navigate = useNavigate();
  const form = useForm<PasswordResetConfirmSchema>({
    defaultValues: { new_password: '' },
    resolver: zodResolver(passwordResetConfirmSchema),
  });

  const mutation = useMutation({
    mutationFn: (values: PasswordResetConfirmSchema) =>
      confirmPasswordReset({ token, new_password: values.new_password }),
  });

  useEffect(() => {
    if (mutation.isSuccess) {
      const timer = setTimeout(() => navigate('/login?password_reset=1'), 2000);
      return () => clearTimeout(timer);
    }
  }, [mutation.isSuccess, navigate]);

  return (
    <form
      className="auth-form"
      onSubmit={form.handleSubmit((values) => {
        mutation.reset();
        void mutation.mutateAsync(values);
      })}
    >
      {mutation.isError ? <ErrorBanner error={mutation.error} /> : null}
      {mutation.isSuccess ? (
        <Notice tone="success">
          Пароль изменён. Перенаправляем на страницу входа…
        </Notice>
      ) : null}

      {showTokenField ? (
        <p>
          Вставьте код из письма (тот же, что в ссылке) или откройте ссылку из письма в браузере.
        </p>
      ) : null}

      {showTokenField ? (
        <Field label="Код сброса пароля">
          <Input
            autoComplete="one-time-code"
            name="password-reset-token"
            onChange={(e) => setManualToken(e.target.value)}
            placeholder="Вставьте код из письма"
            value={manualToken}
          />
        </Field>
      ) : null}

      <Field
        error={form.formState.errors.new_password?.message}
        hint="От 8 до 72 символов: строчные и заглавные буквы, цифра, спецсимвол (!@#$% и др.), без пробелов"
        label="Новый пароль"
      >
        <PasswordInput
          autoComplete="new-password"
          placeholder="Например: Secret1!"
          {...form.register('new_password')}
        />
      </Field>

      <Button
        disabled={mutation.isPending || mutation.isSuccess || !token}
        type="submit"
      >
        {mutation.isPending ? 'Сохраняем...' : 'Установить новый пароль'}
      </Button>

      <div className="auth-form__links">
        <Link to="/password-reset">Запросить новый код</Link>
        <Link to="/login">Вернуться ко входу</Link>
      </div>
    </form>
  );
}
