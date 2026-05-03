import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';

import { listCatalogCategories } from '../../../entities/category/api/categories';
import { listCatalogGifts } from '../../../entities/gift/api/gifts';
import { GiftPreviewCard } from '../../../entities/gift/ui/gift-preview-card';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Container } from '../../../shared/ui/layout/container';

const valueProps = [
  {
    title: 'Быстрый старт',
    text: 'Начните с каталога или короткой анкеты и сразу получите подходящие идеи.',
  },
  {
    title: 'Каталог из разных магазинов',
    text: 'Сравнивайте варианты и переходите к покупке по прямой ссылке.',
  },
  {
    title: 'Подбор под человека',
    text: 'Фильтры и рекомендации помогают искать подарок под повод, интересы и бюджет.',
  },
];

export function HomePage(): JSX.Element {
  const previewGiftsQuery = useQuery({
    queryFn: () =>
      listCatalogGifts({
        has_image: true,
        limit: 4,
        offset: 0,
        sort: 'newest',
      }),
    queryKey: ['catalog', 'preview'],
  });

  const categoriesQuery = useQuery({
    queryFn: () =>
      listCatalogCategories({
        limit: 100,
        offset: 0,
        sort: 'name_asc',
        has_gifts: true,
      }),
    queryKey: ['categories', 'preview', 'with-gifts'],
  });

  return (
    <>
      <section className="hero">
        <Container className="hero__inner">
          <div className="hero__content">
            <p className="eyebrow">Сервис подбора подарков</p>
            <h1>Подобрать подарок без бесконечного скролла маркетплейсов.</h1>
            <p className="hero__description">
              Выберите повод, бюджет или интересы получателя и получите
              подходящие варианты без лишнего поиска.
            </p>
            <div className="hero__actions">
              <Link className={buttonClassName({ size: 'lg' })} to="/recommendation">
                Подобрать подарок
              </Link>
              <Link
                className={buttonClassName({ size: 'lg', variant: 'secondary' })}
                to="/catalog"
              >
                Каталог идей
              </Link>
            </div>
          </div>

          <div className="hero__panel">
            <div className="metric-card">
              <strong>Каталог идей</strong>
              <span>Смотрите подарки по категориям, цене и названию.</span>
            </div>
            <div className="metric-card">
              <strong>Личный кабинет</strong>
              <span>Сохраняйте данные аккаунта и возвращайтесь к подбору позже.</span>
            </div>
            <div className="metric-card">
              <strong>Покупка в пару кликов</strong>
              <span>Открывайте карточку подарка и переходите сразу в магазин.</span>
            </div>
          </div>
        </Container>
      </section>

      <section className="section">
        <Container>
          <div className="section-heading">
            <p className="eyebrow">Почему сервис полезен</p>
            <h2>Интерфейс заточен под сценарий выбора, а не под хаотичную выдачу.</h2>
          </div>
          <div className="value-grid" id="how-it-works">
            {valueProps.map((item) => (
              <article className="value-card" key={item.title}>
                <h3>{item.title}</h3>
                <p>{item.text}</p>
              </article>
            ))}
          </div>
        </Container>
      </section>

      <section className="section section--alt">
        <Container>
          <div className="section-heading section-heading--inline">
            <div>
              <p className="eyebrow">Предпросмотр каталога</p>
              <h2>Несколько свежих идей, чтобы начать без регистрации.</h2>
            </div>
            <Link className={buttonClassName({ variant: 'secondary' })} to="/catalog">
              Открыть каталог
            </Link>
          </div>

          {previewGiftsQuery.isError ? (
            <ErrorBanner
              error={previewGiftsQuery.error}
              title="Не удалось загрузить подборку"
            />
          ) : null}

          {previewGiftsQuery.data?.data.items.length ? (
            <div className="gift-grid">
              {previewGiftsQuery.data.data.items.map((gift) => (
                <GiftPreviewCard gift={gift} key={gift.id} />
              ))}
            </div>
          ) : null}

          {categoriesQuery.data?.data.items.length ? (
            <div className="category-strip">
              {categoriesQuery.data.data.items.map((category) => (
                <span className="chip" key={category.id}>
                  {category.name}
                </span>
              ))}
            </div>
          ) : null}
        </Container>
      </section>
    </>
  );
}
