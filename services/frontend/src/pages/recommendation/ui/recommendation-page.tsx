import { useEffect, useMemo, useState } from 'react';
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

const FALLBACK_IMAGE =
  'https://images.unsplash.com/photo-1513475382585-d06e58bcb0ff?auto=format&fit=crop&w=900&q=80';

const PRESET_OCCASIONS = [
  'День рождения',
  'Юбилей',
  'Новый год',
  '8 Марта',
  '23 Февраля',
  'День всех влюблённых',
  'Годовщина',
  'Свадьба',
  'Выпускной',
  'Рождение ребёнка',
  'Новоселье',
  'Благодарность',
  'Просто так',
] as const;

const PRESET_RELATIONSHIPS = [
  'Партнёру / супругу',
  'Родителям',
  'Ребёнку',
  'Брату или сестре',
  'Другу или подруге',
  'Коллеге',
  'Руководителю',
  'Знакомому',
  'Себе',
] as const;

const PRESET_INTERESTS = [
  'Книги и чтение',
  'Кино и сериалы',
  'Музыка',
  'Спорт и фитнес',
  'Путешествия',
  'Готовка',
  'Настольные и видеоигры',
  'Техника и гаджеты',
  'Рукоделие и DIY',
  'Красота и уход',
  'Автомобили',
  'Фотография',
  'Растения и сад',
  'Дом и уют',
  'Искусство',
] as const;

const RESULTS_PAGE_SIZE = 12;

// ─── Wizard state ────────────────────────────────────────────────────────────

interface WizardData {
  occasion: string;
  relationship: string;
  recipient_age: string;
  recipient_gender: 'male' | 'female' | 'other' | '';
  interest_presets: string[];
  interests_extra: string;
  budget_max: string;
  use_wishlist_context: boolean;
}

const EMPTY: WizardData = {
  occasion: '',
  relationship: '',
  recipient_age: '',
  recipient_gender: '',
  interest_presets: [],
  interests_extra: '',
  budget_max: '',
  use_wishlist_context: true,
};

const STEPS = ['Повод', 'Получатель', 'Интересы', 'Бюджет'];

function mergeInterests(presets: string[], extraCsv: string): string[] {
  const fromExtra = extraCsv
    .split(',')
    .map((s) => s.trim())
    .filter(Boolean);
  return [...new Set([...presets, ...fromExtra])];
}

/** Значение для `<select>` только из пресетов; иначе пусто. */
function valueIfPreset(value: string, presets: readonly string[]): string {
  return (presets as readonly string[]).includes(value) ? value : '';
}

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

  const [resultsPage, setResultsPage] = useState(1);

  // filters for results
  const [filterCategory, setFilterCategory] = useState('');
  const [filterMaxPrice, setFilterMaxPrice] = useState('');

  const track = useTrackEvent();

  const mutation = useMutation({
    mutationFn: createRecommendation,
    onSuccess: (data) => {
      setResultsPage(1);
      track({
        type: 'recommendation_request',
        recommendation_request_id: data.data.recommendation.request_id,
        metadata: { surface: 'recommendation' },
      });
    },
  });

  const requestId = mutation.data?.data.recommendation.request_id;

  const recommendations = useMemo(() => {
    if (!mutation.isSuccess || !mutation.data) return [];
    return mutation.data.data.recommendation.recommendations;
  }, [mutation.isSuccess, mutation.data]);

  const filteredItems = useMemo(
    () =>
      recommendations.filter((item) => {
        if (filterCategory && item.gift.category?.name !== filterCategory) return false;
        if (filterMaxPrice && Number(item.gift.price) > Number(filterMaxPrice)) return false;
        return true;
      }),
    [recommendations, filterCategory, filterMaxPrice],
  );

  const resultsPageCount = Math.max(1, Math.ceil(filteredItems.length / RESULTS_PAGE_SIZE));
  const effectivePage = Math.min(resultsPage, resultsPageCount);

  const paginatedItems = useMemo(() => {
    const start = (effectivePage - 1) * RESULTS_PAGE_SIZE;
    return filteredItems.slice(start, start + RESULTS_PAGE_SIZE);
  }, [filteredItems, effectivePage]);

  // Fire card_view for cards on the current results page
  useEffect(() => {
    if (!mutation.isSuccess) return;
    const start = (effectivePage - 1) * RESULTS_PAGE_SIZE;
    const pageItems = filteredItems.slice(start, start + RESULTS_PAGE_SIZE);
    pageItems.forEach((item) => {
      track({
        type: 'card_view',
        gift_id: item.gift.id,
        recommendation_request_id: requestId,
        metadata: { surface: 'recommendation', position: item.rank },
      });
    });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [mutation.isSuccess, requestId, effectivePage, filteredItems]);

  function update<K extends keyof WizardData>(key: K, value: WizardData[K]) {
    setData((prev) => ({ ...prev, [key]: value }));
    setErrors((prev) => ({ ...prev, [key]: undefined }));
  }

  function togglePresetInterest(label: string) {
    setData((prev) => {
      const has = prev.interest_presets.includes(label);
      const interest_presets = has
        ? prev.interest_presets.filter((x) => x !== label)
        : [...prev.interest_presets, label];
      return { ...prev, interest_presets };
    });
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
      interests: mergeInterests(data.interest_presets, data.interests_extra),
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
    setResultsPage(1);
  }

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
              <Field label="Повод для подарка" hint="Выберите из списка">
                <select
                  className="input"
                  value={valueIfPreset(data.occasion, PRESET_OCCASIONS)}
                  onChange={(e) => update('occasion', e.target.value)}
                >
                  <option value="">Не выбрано</option>
                  {PRESET_OCCASIONS.map((o) => (
                    <option key={o} value={o}>
                      {o}
                    </option>
                  ))}
                </select>
              </Field>
              <Field label="Кому дарите" hint="Выберите из списка">
                <select
                  className="input"
                  value={valueIfPreset(data.relationship, PRESET_RELATIONSHIPS)}
                  onChange={(e) => update('relationship', e.target.value)}
                >
                  <option value="">Не выбрано</option>
                  {PRESET_RELATIONSHIPS.map((r) => (
                    <option key={r} value={r}>
                      {r}
                    </option>
                  ))}
                </select>
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
                hint="Нажмите, чтобы отметить один или несколько вариантов"
              >
                <div
                  className="wizard__interest-chips"
                  role="group"
                  aria-label="Типичные интересы"
                >
                  {PRESET_INTERESTS.map((label) => {
                    const selected = data.interest_presets.includes(label);
                    return (
                      <button
                        key={label}
                        type="button"
                        className={[
                          'chip',
                          'wizard__interest-chip',
                          selected ? 'wizard__interest-chip--selected' : '',
                        ].join(' ')}
                        aria-pressed={selected}
                        onClick={() => togglePresetInterest(label)}
                      >
                        {label}
                      </button>
                    );
                  })}
                </div>
              </Field>
              <Field
                label="Дополнительно"
                hint="Свои интересы через запятую, если их нет в списке выше"
              >
                <Input
                  placeholder="Например: винил, настолки"
                  value={data.interests_extra}
                  onChange={(e) => update('interests_extra', e.target.value)}
                />
              </Field>
            </div>
          )}

          {/* ── Step 3: Бюджет ── */}
          {step === 3 && (
            <div className="wizard__body">
              <h2 className="wizard__title">Бюджет и параметры</h2>
              <Field label="Бюджет до" error={errors.budget_max}>
                <Input
                  placeholder="Например: 5000"
                  value={data.budget_max}
                  onChange={(e) => update('budget_max', e.target.value)}
                />
              </Field>
              <label className="checkbox-field">
                <input
                  className="checkbox-field__native"
                  type="checkbox"
                  checked={data.use_wishlist_context}
                  onChange={(e) => update('use_wishlist_context', e.target.checked)}
                />
                <span className="checkbox-field__indicator" aria-hidden />
                <span className="checkbox-field__body">
                  <span className="checkbox-field__title">Учитывать список желаний</span>
                  <span className="checkbox-field__hint">
                    Подбор может опираться на подарки, которые вы сохранили в списке.
                  </span>
                </span>
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
            <span>
              Подобрано: {recommendations.length}
              {filteredItems.length !== recommendations.length
                ? ` (после фильтров: ${filteredItems.length})`
                : null}
            </span>
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
                    onChange={(e) => {
                      setFilterCategory(e.target.value);
                      setResultsPage(1);
                    }}
                  >
                    <option value="">Все категории</option>
                    {categories.map((c) => (
                      <option key={c} value={c}>{c}</option>
                    ))}
                  </select>
                  <Input
                    placeholder="Цена до"
                    style={{ width: '120px' }}
                    type="number"
                    value={filterMaxPrice}
                    onChange={(e) => {
                      setFilterMaxPrice(e.target.value);
                      setResultsPage(1);
                    }}
                  />
                  {(filterCategory || filterMaxPrice) && (
                    <button
                      className={buttonClassName({ variant: 'ghost' })}
                      type="button"
                      onClick={() => {
                        setFilterCategory('');
                        setFilterMaxPrice('');
                        setResultsPage(1);
                      }}
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
                <>
                  <div className="gift-grid">
                    {paginatedItems.map((item) => (
                      <article className="card gift-card" key={item.gift.id}>
                      <Link className="gift-card__image-link" to={`/catalog/${item.gift.id}`}>
                        <img
                          alt={item.gift.name}
                          className="gift-card__image"
                          src={item.gift.image ?? FALLBACK_IMAGE}
                        />
                      </Link>
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

                        {/* Alternatives */}
                        {item.alternatives.length > 0 && (
                          <details className="recommendation-alternatives">
                            <summary>
                              Похожие варианты ({item.alternatives.length})
                            </summary>
                            <ul className="recommendation-alternatives__list">
                              {item.alternatives.map((alt) => (
                                <li
                                  className="recommendation-alternatives__item"
                                  key={alt.gift.id}
                                >
                                  <div className="recommendation-alternatives__info">
                                    <span>{alt.gift.name}</span>
                                    <strong>{formatPrice(alt.gift.price)}</strong>
                                  </div>
                                  <Link
                                    className={buttonClassName()}
                                    to={`/catalog/${alt.gift.id}`}
                                  >
                                    Подробнее
                                  </Link>
                                </li>
                              ))}
                            </ul>
                          </details>
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

                  {resultsPageCount > 1 && (
                    <nav
                      className="recommendation-pagination"
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        gap: '1rem',
                        flexWrap: 'wrap',
                        marginTop: '1.25rem',
                      }}
                      aria-label="Страницы результатов"
                    >
                      <Button
                        variant="ghost"
                        type="button"
                        disabled={effectivePage <= 1}
                        onClick={() => setResultsPage((p) => Math.max(1, p - 1))}
                      >
                        Назад
                      </Button>
                      <span style={{ fontSize: '0.9rem', color: 'var(--color-muted)' }}>
                        Страница {effectivePage} из {resultsPageCount}
                      </span>
                      <Button
                        variant="ghost"
                        type="button"
                        disabled={effectivePage >= resultsPageCount}
                        onClick={() => setResultsPage((p) => Math.min(resultsPageCount, p + 1))}
                      >
                        Вперёд
                      </Button>
                    </nav>
                  )}
                </>
              )}
            </>
          )}
        </>
      )}
    </Container>
  );
}
