import { useEffect, useState } from 'react';
import { useMutation } from '@tanstack/react-query';
import { Link } from 'react-router-dom';

import { createRecommendation } from '../../../features/recommendation/api/recommendation';
import { useTrackEvent } from '../../../features/tracking/model/use-track-event';
import { WishlistSaveButton } from '../../../features/wishlist/ui/wishlist-save-button';
import { formatPrice } from '../../../shared/lib/format';
import { Button } from '../../../shared/ui/button/button';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { EmptyState } from '../../../shared/ui/feedback/empty-state';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Notice } from '../../../shared/ui/feedback/notice';
import { Field } from '../../../shared/ui/form/field';
import { Input } from '../../../shared/ui/input/input';
import { Container } from '../../../shared/ui/layout/container';

// ─── Wizard state ────────────────────────────────────────────────────────────

interface WizardData {
  occasion: string;
  relationship: string;
  recipient_age: string;
  recipient_gender: 'male' | 'female' | 'other' | '';
  interests: string;
  budget_max: string;
  top_n: number;
  use_wishlist_context: boolean;
}

const EMPTY: WizardData = {
  occasion: '',
  relationship: '',
  recipient_age: '',
  recipient_gender: '',
  interests: '',
  budget_max: '',
  top_n: 5,
  use_wishlist_context: true,
};

const STEPS = ['Повод', 'Получатель', 'Интересы', 'Бюджет'];

// ─── Helper: simple numeric validation ───────────────────────────────────────

function validateAge(v: string): string | null {
  if (v === '') return null;
  const n = Number(v);
  if (!Number.isInteger(n) || n < 0 || n > 120) return 'Введите возраст от 0 до 120';
  return null;
}

function validateBudget(v: string): string | null {
  if (!v.trim()) return 'Укажите бюджет';
  if (isNaN(Number(v)) || Number(v) <= 0) return 'Введите положительное число';
  return null;
}

// ─── Page ─────────────────────────────────────────────────────────────────────

export function RecommendationPage(): JSX.Element {
  const [step, setStep] = useState(0);
  const [data, setData] = useState<WizardData>(EMPTY);
  const [errors, setErrors] = useState<Partial<Record<keyof WizardData, string>>>({});

  // filters for results
  const [filterCategory, setFilterCategory] = useState('');
  const [filterMaxPrice, setFilterMaxPrice] = useState('');

  const mutation = useMutation({ mutationFn: createRecommendation });
  const track = useTrackEvent();

  const requestId = mutation.data?.data.recommendation.request_id;

  // Fire card_view for every shown recommendation card
  useEffect(() => {
    if (!mutation.isSuccess) return;
    const items = mutation.data.data.recommendation.recommendations;
    items.forEach((item) => {
      track({
        type: 'card_view',
        gift_id: item.gift.id,
        recommendation_request_id: requestId,
        metadata: { surface: 'recommendation', position: item.rank },
      });
    });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mutation.isSuccess, requestId]);

  function update<K extends keyof WizardData>(key: K, value: WizardData[K]) {
    setData((prev) => ({ ...prev, [key]: value }));
    setErrors((prev) => ({ ...prev, [key]: undefined }));
  }

  function validateStep(): boolean {
    const next: typeof errors = {};
    if (step === 1) {
      const ageErr = validateAge(data.recipient_age);
      if (ageErr) next.recipient_age = ageErr;
    }
    if (step === 3) {
      const budgetErr = validateBudget(data.budget_max);
      if (budgetErr) next.budget_max = budgetErr;
    }
    setErrors(next);
    return Object.keys(next).length === 0;
  }

  function handleNext() {
    if (!validateStep()) return;
    if (step < STEPS.length - 1) {
      setStep((s) => s + 1);
    } else {
      submit();
    }
  }

  function submit() {
    mutation.reset();
    void mutation.mutateAsync({
      occasion: data.occasion.trim() || undefined,
      relationship: data.relationship.trim() || undefined,
      recipient_age: data.recipient_age.trim() ? Number(data.recipient_age) : undefined,
      recipient_gender: (data.recipient_gender || undefined) as 'male' | 'female' | 'other' | undefined,
      budget_max: data.budget_max.trim(),
      interests: data.interests
        .split(',')
        .map((s) => s.trim())
        .filter(Boolean),
      top_n: data.top_n,
      use_wishlist_context: data.use_wishlist_context,
    });
  }

  function restart() {
    setData(EMPTY);
    setErrors({});
    setStep(0);
    setFilterCategory('');
    setFilterMaxPrice('');
    mutation.reset();
  }

  const recommendations = mutation.data?.data.recommendation.recommendations ?? [];

  const filteredItems = recommendations.filter((item) => {
    if (filterCategory && item.gift.category?.name !== filterCategory) return false;
    if (filterMaxPrice && Number(item.gift.price) > Number(filterMaxPrice)) return false;
    return true;
  });

  const categories = Array.from(
    new Set(recommendations.map((i) => i.gift.category?.name).filter(Boolean)),
  ) as string[];

  const isLastStep = step === STEPS.length - 1;

  return (
    <Container className="page-stack">
      <section className="section-heading">
        <p className="eyebrow">Подбор подарка</p>
        <h1>Мастер подбора подарка</h1>
        <p className="page-copy">
          Расскажите немного о получателе, и мы подберём подходящие варианты.
        </p>
      </section>

      {/* ── Wizard (hidden after successful submit) ── */}
      {!mutation.isSuccess && (
        <div className="wizard">
          {/* Step indicator */}
          <div className="wizard__steps">
            {STEPS.map((label, i) => (
              <button
                className={[
                  'wizard__step',
                  i === step ? 'wizard__step--active' : '',
                  i < step ? 'wizard__step--done' : '',
                ].join(' ')}
                disabled={i > step}
                key={label}
                onClick={() => { if (i < step) setStep(i); }}
                type="button"
              >
                <span className="wizard__step-number">{i + 1}</span>
                <span className="wizard__step-label">{label}</span>
              </button>
            ))}
          </div>

          {/* ── Step 0: Повод и отношения ── */}
          {step === 0 && (
            <div className="wizard__body">
              <h2 className="wizard__title">Повод и отношения</h2>
              <Field label="Повод для подарка" hint="Необязательно">
                <Input
                  placeholder="Например: день рождения"
                  value={data.occasion}
                  onChange={(e) => update('occasion', e.target.value)}
                />
              </Field>
              <Field label="Кому дарите" hint="Необязательно">
                <Input
                  placeholder="Например: коллега, друг, родственник"
                  value={data.relationship}
                  onChange={(e) => update('relationship', e.target.value)}
                />
              </Field>
            </div>
          )}

          {/* ── Step 1: О получателе ── */}
          {step === 1 && (
            <div className="wizard__body">
              <h2 className="wizard__title">О получателе</h2>
              <Field label="Возраст" hint="Необязательно" error={errors.recipient_age}>
                <Input
                  type="number"
                  min={0}
                  max={120}
                  placeholder="Например: 25"
                  value={data.recipient_age}
                  onChange={(e) => update('recipient_age', e.target.value)}
                />
              </Field>
              <Field label="Пол" hint="Необязательно">
                <select
                  className="input"
                  value={data.recipient_gender}
                  onChange={(e) =>
                    update('recipient_gender', e.target.value as WizardData['recipient_gender'])
                  }
                >
                  <option value="">Не указан</option>
                  <option value="male">Мужской</option>
                  <option value="female">Женский</option>
                  <option value="other">Другой</option>
                </select>
              </Field>
            </div>
          )}

          {/* ── Step 2: Интересы ── */}
          {step === 2 && (
            <div className="wizard__body">
              <h2 className="wizard__title">Интересы получателя</h2>
              <Field
                label="Интересы"
                hint="Через запятую: спорт, книги, музыка"
              >
                <Input
                  placeholder="Интересы получателя"
                  value={data.interests}
                  onChange={(e) => update('interests', e.target.value)}
                />
              </Field>
            </div>
          )}

          {/* ── Step 3: Бюджет ── */}
          {step === 3 && (
            <div className="wizard__body">
              <h2 className="wizard__title">Бюджет и параметры</h2>
              <Field label="Бюджет до, ₽" error={errors.budget_max}>
                <Input
                  placeholder="Например: 5000"
                  value={data.budget_max}
                  onChange={(e) => update('budget_max', e.target.value)}
                />
              </Field>
              <Field label="Сколько вариантов показать">
                <Input
                  type="number"
                  min={1}
                  max={20}
                  value={data.top_n}
                  onChange={(e) => update('top_n', Number(e.target.value))}
                />
              </Field>
              <label className="field">
                <span className="field__label">Учитывать список желаний</span>
                <input
                  type="checkbox"
                  checked={data.use_wishlist_context}
                  onChange={(e) => update('use_wishlist_context', e.target.checked)}
                />
              </label>
            </div>
          )}

          {mutation.isError && (
            <ErrorBanner error={mutation.error} title="Не удалось получить рекомендации" />
          )}

          {/* Navigation */}
          <div className="wizard__nav">
            {step > 0 && (
              <Button variant="ghost" type="button" onClick={() => setStep((s) => s - 1)}>
                Назад
              </Button>
            )}
            <Button disabled={mutation.isPending} type="button" onClick={handleNext}>
              {mutation.isPending
                ? 'Подбираем…'
                : isLastStep
                  ? 'Подобрать подарок'
                  : 'Далее'}
            </Button>
            {step < STEPS.length - 1 && (
              <button
                className={buttonClassName({ variant: 'ghost' })}
                type="button"
                onClick={() => setStep((s) => s + 1)}
              >
                Пропустить
              </button>
            )}
          </div>
        </div>
      )}

      {/* ── Results ── */}
      {mutation.isSuccess && (
        <>
          <div className="catalog-summary" style={{ display: 'flex', gap: '1rem', alignItems: 'center', flexWrap: 'wrap' }}>
            <span>Подобрано: {recommendations.length}</span>
            <Button variant="ghost" type="button" onClick={restart}>
              Начать заново
            </Button>
          </div>

          {recommendations.length === 0 ? (
            <EmptyState
              description="Попробуйте увеличить бюджет или уточнить интересы."
              title="По выбранным критериям ничего не найдено"
            />
          ) : (
            <>
              {mutation.data.data.recommendation.fallback_used && (
                <Notice>Мы подобрали варианты на основе доступных данных.</Notice>
              )}

              {/* Filters */}
              {categories.length > 0 && (
                <div className="catalog-filters" style={{ display: 'flex', gap: '0.5rem', flexWrap: 'wrap', alignItems: 'center' }}>
                  <span style={{ fontSize: '0.875rem', color: 'var(--color-muted)' }}>Фильтры:</span>
                  <select
                    className="input"
                    style={{ width: 'auto' }}
                    value={filterCategory}
                    onChange={(e) => setFilterCategory(e.target.value)}
                  >
                    <option value="">Все категории</option>
                    {categories.map((c) => (
                      <option key={c} value={c}>{c}</option>
                    ))}
                  </select>
                  <Input
                    placeholder="Цена до, ₽"
                    style={{ width: '120px' }}
                    type="number"
                    value={filterMaxPrice}
                    onChange={(e) => setFilterMaxPrice(e.target.value)}
                  />
                  {(filterCategory || filterMaxPrice) && (
                    <button
                      className={buttonClassName({ variant: 'ghost' })}
                      type="button"
                      onClick={() => { setFilterCategory(''); setFilterMaxPrice(''); }}
                    >
                      Сбросить
                    </button>
                  )}
                </div>
              )}

              {filteredItems.length === 0 ? (
                <EmptyState
                  description="Попробуйте изменить фильтры."
                  title="Ничего не подходит под фильтры"
                />
              ) : (
                <div className="gift-grid">
                  {filteredItems.map((item) => (
                    <article className="card gift-card" key={item.gift.id}>
                      <div className="gift-card__content">
                        <div className="gift-card__heading">
                          <h3>{item.rank}. {item.gift.name}</h3>
                          <strong>{formatPrice(item.gift.price)}</strong>
                        </div>
                        <p>{item.gift.description}</p>

                        {/* Explanations */}
                        {item.explanations.length > 0 && (
                          <ul className="recommendation-explanations">
                            {item.explanations.map((exp) => (
                              <li key={exp.code} className="recommendation-explanation">
                                ✓ {exp.text}
                              </li>
                            ))}
                          </ul>
                        )}

                        <div className="gift-card__actions">
                          <Link className={buttonClassName()} to={`/catalog/${item.gift.id}`}>
                            Подробнее
                          </Link>
                          <WishlistSaveButton giftID={item.gift.id} />
                          <a
                            className={buttonClassName({ variant: 'ghost' })}
                            href={item.gift.store_link}
                            rel="noreferrer"
                            target="_blank"
                            onClick={() =>
                              track({
                                type: 'outbound_click',
                                gift_id: item.gift.id,
                                recommendation_request_id: requestId,
                                metadata: { surface: 'recommendation', position: item.rank },
                              })
                            }
                          >
                            В магазин
                          </a>
                        </div>
                      </div>
                    </article>
                  ))}
                </div>
              )}
            </>
          )}
        </>
      )}
    </Container>
  );
}
