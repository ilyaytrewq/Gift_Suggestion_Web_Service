import { zodResolver } from '@hookform/resolvers/zod';
import { useMutation } from '@tanstack/react-query';
import { useForm } from 'react-hook-form';
import { Link } from 'react-router-dom';
import { z } from 'zod';

import { createRecommendation } from '../../../features/recommendation/api/recommendation';
import { formatPrice } from '../../../shared/lib/format';
import { Button } from '../../../shared/ui/button/button';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { EmptyState } from '../../../shared/ui/feedback/empty-state';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';
import { Container } from '../../../shared/ui/layout/container';

const recommendationSchema = z.object({
  occasion: z.string().max(120, 'Слишком длинное значение').optional(),
  relationship: z.string().max(120, 'Слишком длинное значение').optional(),
  recipient_age: z
    .string()
    .trim()
    .refine(
      (value) =>
        value === '' || (/^\d+$/.test(value) && Number(value) <= 120),
      'Введите возраст числом от 0 до 120',
    )
    .optional(),
  budget_max: z.string().trim().min(1, 'Укажите верхний бюджет'),
  interests: z.string().optional(),
  top_n: z.number().int().min(1, 'Минимум 1').max(20, 'Максимум 20'),
  use_wishlist_context: z.boolean(),
});

type RecommendationSchema = z.infer<typeof recommendationSchema>;

export function RecommendationPage(): JSX.Element {
  const form = useForm<RecommendationSchema>({
    defaultValues: {
      occasion: '',
      relationship: '',
      recipient_age: '',
      budget_max: '',
      interests: '',
      top_n: 5,
      use_wishlist_context: true,
    },
    resolver: zodResolver(recommendationSchema),
  });

  const mutation = useMutation({
    mutationFn: createRecommendation,
  });

  return (
    <Container className="page-stack">
      <section className="section-heading">
        <p className="eyebrow">Recommendation Wizard</p>
        <h1>Мастер подбора подарка</h1>
        <p className="page-copy">
          Заполните параметры и получите персональные рекомендации. Открытие
          и отправка анкеты доступны без логина. Авторизация остаётся полезной
          для персонализации по wishlist-контексту.
        </p>
      </section>

      <form
        className="profile-form"
        onSubmit={form.handleSubmit((values) => {
          mutation.reset();
          void mutation.mutateAsync({
            occasion: values.occasion?.trim() || undefined,
            relationship: values.relationship?.trim() || undefined,
            recipient_age: values.recipient_age?.trim()
              ? Number(values.recipient_age)
              : undefined,
            budget_max: values.budget_max.trim(),
            interests: values.interests
              ?.split(',')
              .map((interest: string) => interest.trim())
              .filter(Boolean),
            top_n: values.top_n,
            use_wishlist_context: values.use_wishlist_context,
          });
        })}
      >
        <Field error={form.formState.errors.occasion?.message} label="Повод">
          <Input
            placeholder="Например: день рождения"
            {...form.register('occasion')}
          />
        </Field>

        <Field error={form.formState.errors.relationship?.message} label="Кем вам приходится получатель">
          <Input
            placeholder="Например: коллега, друг, родственник"
            {...form.register('relationship')}
          />
        </Field>

        <Field error={form.formState.errors.recipient_age?.message} hint="Необязательно" label="Возраст получателя">
          <Input
            min={0}
            placeholder="Например: 25"
            type="number"
            {...form.register('recipient_age')}
          />
        </Field>

        <Field error={form.formState.errors.budget_max?.message} label="Бюджет до">
          <Input
            placeholder="Например: 5000"
            {...form.register('budget_max')}
          />
        </Field>

        <Field
          error={form.formState.errors.interests?.message}
          hint="Через запятую, например: спорт, технологии, музыка"
          label="Интересы"
        >
          <Input
            placeholder="Интересы получателя"
            {...form.register('interests')}
          />
        </Field>

        <Field error={form.formState.errors.top_n?.message} label="Сколько рекомендаций показать">
          <Input
            max={20}
            min={1}
            type="number"
            {...form.register('top_n', { valueAsNumber: true })}
          />
        </Field>

        <label className="field">
          <span className="field__label">Учитывать wishlist-контекст</span>
          <input type="checkbox" {...form.register('use_wishlist_context')} />
        </label>

        {mutation.isError ? (
          <ErrorBanner error={mutation.error} title="Не удалось получить рекомендации" />
        ) : null}

        <div className="profile-actions">
          <Button disabled={mutation.isPending} type="submit">
            {mutation.isPending ? 'Подбираем...' : 'Подобрать подарок'}
          </Button>
          <Link className={buttonClassName({ variant: 'ghost' })} to="/catalog">
            Перейти в каталог
          </Link>
        </div>
      </form>

      {mutation.isSuccess ? (
        mutation.data.data.recommendation.recommendations.length ? (
          <>
            <div className="catalog-summary">
              <span>
                Статус: {mutation.data.data.recommendation.status}
              </span>
              <span>
                Подобрано: {mutation.data.data.recommendation.recommendations.length}
              </span>
            </div>
            <div className="gift-grid">
              {mutation.data.data.recommendation.recommendations.map((item) => (
                <article className="card gift-card" key={item.gift.id}>
                  <div className="gift-card__content">
                    <div className="gift-card__heading">
                      <h3>{item.rank}. {item.gift.name}</h3>
                      <strong>{formatPrice(item.gift.price)}</strong>
                    </div>
                    <p>{item.gift.description}</p>
                    <div className="gift-card__actions">
                      <Link className={buttonClassName()} to={`/catalog/${item.gift.id}`}>
                        Открыть карточку
                      </Link>
                      <a
                        className={buttonClassName({ variant: 'ghost' })}
                        href={item.gift.store_link}
                        rel="noreferrer"
                        target="_blank"
                      >
                        Купить
                      </a>
                    </div>
                  </div>
                </article>
              ))}
            </div>
          </>
        ) : (
          <EmptyState
            description="Попробуйте увеличить бюджет или уточнить интересы."
            title="По выбранным критериям ничего не найдено"
          />
        )
      ) : null}
    </Container>
  );
}
