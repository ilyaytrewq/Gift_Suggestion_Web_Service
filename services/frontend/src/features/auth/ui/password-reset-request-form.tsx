import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';

import { requestPasswordReset } from '../api/auth';
import {
  passwordResetRequestSchema,
  type PasswordResetRequestSchema,
} from '../model/schemas';
import { Button } from '../../../shared/ui/button/button';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';

export function PasswordResetRequestForm(): JSX.Element {
  const form = useForm<PasswordResetRequestSchema>({
    defaultValues: {
      email: '',
    },
    resolver: zodResolver(passwordResetRequestSchema),
  });

  const mutation = useMutation({
    mutationFn: requestPasswordReset,
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
          Если аккаунт с таким email существует, письмо на восстановление уже
          запрошено.
        </Notice>
      ) : null}

      <Field
        error={form.formState.errors.email?.message}
        hint="Backend принимает только email и отвечает статусом accepted."
        label="Email"
      >
        <Input
          autoComplete="email"
          placeholder="you@example.com"
          {...form.register('email')}
        />
      </Field>

      <Button disabled={mutation.isPending} type="submit">
        {mutation.isPending ? 'Отправляем...' : 'Запросить восстановление'}
      </Button>

      <div className="auth-form__links">
        <Link to="/login">Вернуться ко входу</Link>
      </div>
    </form>
  );
}
