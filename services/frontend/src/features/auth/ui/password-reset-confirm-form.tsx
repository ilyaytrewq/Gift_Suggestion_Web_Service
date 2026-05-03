import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';

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

interface Props {
  token: string;
}

export function PasswordResetConfirmForm({ token }: Props): JSX.Element {
  const form = useForm<PasswordResetConfirmSchema>({
    defaultValues: { new_password: '' },
    resolver: zodResolver(passwordResetConfirmSchema),
  });

  const mutation = useMutation({
    mutationFn: (values: PasswordResetConfirmSchema) =>
      confirmPasswordReset({ token, new_password: values.new_password }),
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
      {mutation.isSuccess ? (
        <Notice tone="success">
          Пароль изменён. Теперь можно{' '}
          <Link to="/login">войти</Link> с новым паролем.
        </Notice>
      ) : null}

      <Field
        error={form.formState.errors.new_password?.message}
        label="Новый пароль"
      >
        <Input
          autoComplete="new-password"
          placeholder="Минимум 8 символов"
          type="password"
          {...form.register('new_password')}
        />
      </Field>

      <Button disabled={mutation.isPending || mutation.isSuccess} type="submit">
        {mutation.isPending ? 'Сохраняем...' : 'Установить новый пароль'}
      </Button>

      <div className="auth-form__links">
        <Link to="/login">Вернуться ко входу</Link>
      </div>
    </form>
  );
}
