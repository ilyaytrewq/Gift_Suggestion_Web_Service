import { zodResolver } from '@hookform/resolvers/zod';
import {
  useMutation,
  useQuery,
} from '@tanstack/react-query';
import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';
import { z } from 'zod';

import {
  getCurrentUser,
  updateCurrentUser,
} from '../../../features/auth/api/auth';
import { useAuth } from '../../../shared/auth/use-auth';
import { formatDateTime } from '../../../shared/lib/format';
import { Button } from '../../../shared/ui/button/button';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { PageLoader } from '../../../shared/ui/feedback/page-loader';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';
import { Container } from '../../../shared/ui/layout/container';

const profileSchema = z.object({
  display_name: z
    .string()
    .max(120, 'Имя должно быть не длиннее 120 символов'),
});

type ProfileSchema = z.infer<typeof profileSchema>;

export function ProfilePage(): JSX.Element {
  const auth = useAuth();
  const form = useForm<ProfileSchema>({
    defaultValues: {
      display_name: auth.user?.display_name ?? '',
    },
    resolver: zodResolver(profileSchema),
  });

  const query = useQuery({
    queryKey: ['current-user'],
    queryFn: getCurrentUser,
    enabled: Boolean(auth.accessToken),
    initialData: auth.user
      ? {
        status: 'ok' as const,
        data: { user: auth.user },
        meta: { request_id: 'local-auth-context' },
      }
      : undefined,
  });

  const user = query.data?.data.user ?? null;

  useEffect(() => {
    if (user) {
      form.reset({ display_name: user.display_name });
      auth.setCurrentUser(user);
    }
  }, [auth, form, user]);

  const mutation = useMutation({
    mutationFn: updateCurrentUser,
    onSuccess: (payload) => {
      auth.setCurrentUser(payload.data.user);
      form.reset({ display_name: payload.data.user.display_name });
    },
  });

  if (!auth.accessToken) {
    return (
      <Container className="profile-page">
        <section className="profile-card">
          <p className="eyebrow">Account</p>
          <h1>Личные данные доступны после входа.</h1>
          <p className="page-copy">
            Войдите в аккаунт, чтобы открыть профиль, проверить email и изменить имя.
          </p>
          <Link className={buttonClassName()} to="/login">
            Войти
          </Link>
        </section>
      </Container>
    );
  }

  if (query.isPending) {
    return (
      <PageLoader
        title="Загружаем профиль"
        description="Получаем актуальные данные аккаунта."
      />
    );
  }

  if (query.isError) {
    return (
      <Container className="profile-page">
        <ErrorBanner
          error={query.error}
          title="Не удалось загрузить личные данные"
        />
      </Container>
    );
  }

  if (!user) {
    return (
      <Container className="profile-page">
        <Notice>
          Профиль пока недоступен. Попробуйте обновить страницу или войти заново.
        </Notice>
      </Container>
    );
  }

  return (
    <Container className="profile-page">
      <section className="profile-card">
        <div className="profile-card__header">
          <div>
            <p className="eyebrow">Account</p>
            <h1>Личные данные</h1>
            <p className="page-copy">
              Здесь можно проверить данные аккаунта и обновить отображаемое имя.
            </p>
          </div>
          <span className="chip chip--muted">{user.role}</span>
        </div>

        <dl className="profile-list">
          <div>
            <dt>Email</dt>
            <dd>{user.email}</dd>
          </div>
          <div>
            <dt>Создан</dt>
            <dd>{formatDateTime(user.created_at)}</dd>
          </div>
          <div>
            <dt>Последний вход</dt>
            <dd>{formatDateTime(user.last_login_at)}</dd>
          </div>
        </dl>

        <form
          className="profile-form"
          onSubmit={form.handleSubmit((values) => {
            mutation.reset();
            void mutation.mutateAsync({
              profile: {
                display_name: values.display_name.trim() || null,
              },
            });
          })}
        >
          {mutation.isError ? (
            <ErrorBanner
              error={mutation.error}
              title="Не удалось обновить профиль"
            />
          ) : null}
          {mutation.isSuccess ? (
            <Notice tone="success">Профиль обновлён.</Notice>
          ) : null}

          <Field
            error={form.formState.errors.display_name?.message}
            hint="Можно оставить пустым"
            label="Имя"
          >
            <Input
              autoComplete="name"
              placeholder="Как к вам обращаться"
              {...form.register('display_name')}
            />
          </Field>

          <div className="profile-actions">
            <Button disabled={mutation.isPending} type="submit">
              {mutation.isPending ? 'Сохраняем...' : 'Сохранить'}
            </Button>
            <Link className={buttonClassName({ variant: 'ghost' })} to="/catalog">
              Вернуться в каталог
            </Link>
          </div>
        </form>
      </section>
    </Container>
  );
}
