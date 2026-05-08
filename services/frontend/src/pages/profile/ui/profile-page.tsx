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
  promoteUserToAdmin,
  updateCurrentUser,
} from '../../../features/auth/api/auth';
import { VkConnectionPanel } from '../../../features/vk-integration/ui/vk-connection-panel';
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

const adminPromoteSchema = z.object({
  email: z
    .string()
    .trim()
    .min(1, 'Укажите email')
    .email('Введите корректный адрес'),
});

type ProfileSchema = z.infer<typeof profileSchema>;
type AdminPromoteSchema = z.infer<typeof adminPromoteSchema>;

function getRoleLabel(role: string): string | null {
  if (role === 'admin') {
    return 'Администратор';
  }

  return null;
}

export function ProfilePage(): JSX.Element {
  const auth = useAuth();
  const form = useForm<ProfileSchema>({
    defaultValues: {
      display_name: auth.user?.display_name ?? '',
    },
    resolver: zodResolver(profileSchema),
  });

  const promoteForm = useForm<AdminPromoteSchema>({
    defaultValues: {
      email: '',
    },
    resolver: zodResolver(adminPromoteSchema),
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

  const promoteMutation = useMutation({
    mutationFn: (payload: AdminPromoteSchema) => promoteUserToAdmin({
      email: payload.email.trim(),
    }),
    onSuccess: () => {
      promoteForm.reset({ email: '' });
    },
  });

  if (!auth.accessToken) {
    return (
      <Container className="profile-page">
        <section className="profile-card">
          <p className="eyebrow">Профиль</p>
          <h1>Войдите, чтобы открыть профиль.</h1>
          <p className="page-copy">
            После входа здесь можно проверить email и изменить имя.
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
        description="Обновляем данные аккаунта."
      />
    );
  }

  if (query.isError) {
    return (
      <Container className="profile-page">
        <ErrorBanner
          error={query.error}
          title="Не удалось открыть профиль"
        />
      </Container>
    );
  }

  if (!user) {
    return (
      <Container className="profile-page">
        <Notice>
          Не удалось загрузить профиль. Попробуйте войти ещё раз.
        </Notice>
      </Container>
    );
  }

  return (
    <Container className="profile-page page-stack">
      <section className="profile-card">
        <div className="profile-card__header">
          <div>
            <p className="eyebrow">Профиль</p>
            <h1>Личные данные</h1>
            <p className="page-copy">
              Проверьте данные аккаунта и обновите имя, если нужно.
            </p>
          </div>
          {getRoleLabel(user.role) ? (
            <span className="chip chip--muted">{getRoleLabel(user.role)}</span>
          ) : null}
        </div>

        <dl className="profile-list">
          <div>
            <dt>Электронная почта</dt>
            <dd>{user.email}</dd>
          </div>
          <div>
            <dt>Дата регистрации</dt>
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
              title="Не удалось сохранить изменения"
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

      <VkConnectionPanel />

      {user.role === 'admin' ? (
        <section className="profile-card">
          <div className="profile-card__header">
            <div>
              <p className="eyebrow">Администрирование</p>
              <h2>Сделать администратором</h2>
              <p className="page-copy">
                Укажите электронную почту уже зарегистрированного пользователя. После сохранения
                ему нужно выйти и войти снова, чтобы токен обновился.
              </p>
            </div>
          </div>

          <form
            className="profile-form"
            onSubmit={promoteForm.handleSubmit((values) => {
              promoteMutation.reset();
              void promoteMutation.mutateAsync(values);
            })}
          >
            {promoteMutation.isError ? (
              <ErrorBanner
                error={promoteMutation.error}
                title="Не удалось назначить роль"
              />
            ) : null}
            {promoteMutation.isSuccess && promoteMutation.data ? (
              <Notice tone="success">
                Роль назначена: {promoteMutation.data.data.user.email}. Пользователь должен заново войти в аккаунт.
              </Notice>
            ) : null}

            <Field
              error={promoteForm.formState.errors.email?.message}
              hint="Учётная запись с этим адресом должна уже существовать."
              label="Email пользователя"
            >
              <Input
                autoComplete="email"
                placeholder="partner@example.com"
                type="email"
                {...promoteForm.register('email')}
              />
            </Field>

            <div className="profile-actions">
              <Button disabled={promoteMutation.isPending} type="submit">
                {promoteMutation.isPending ? 'Назначаем...' : 'Назначить администратора'}
              </Button>
              <Link className={buttonClassName({ variant: 'ghost' })} to="/admin/import">
                Импорт каталога
              </Link>
            </div>
          </form>
        </section>
      ) : null}
    </Container>
  );
}
