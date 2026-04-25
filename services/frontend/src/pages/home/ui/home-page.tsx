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
    title: 'Объяснимые рекомендации',
    text: 'В следующих срезах мастер подбора будет выдавать не просто список, а понятные причины выбора.',
  },
  {
    title: 'Каталог из разных магазинов',
    text: 'Уже сейчас можно просматривать идеи и уходить на покупку по реальной ссылке магазина.',
  },
  {
    title: 'Подготовка под wishlist',
    text: 'Структура Slice 1 уже учитывает будущую работу со списками желаний без переделки catalog flow.',
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
        limit: 6,
        offset: 0,
        sort: 'name_asc',
      }),
    queryKey: ['categories', 'preview'],
  });

  return (
    <>
      <section className="hero">
        <Container className="hero__inner">
          <div className="hero__content">
            <p className="eyebrow">Gift Suggestion Web Service</p>
            <h1>Подобрать подарок без бесконечного скролла маркетплейсов.</h1>
            <p className="hero__description">
              Slice 1 запускает public MVP: каталог идей, карточки подарков и
              auth foundation. Персональный мастер подбора уже доступен: открыть
              его можно по кнопке ниже.
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
              <strong>Каталог</strong>
              <span>Поиск, категории и карточки подарков уже доступны.</span>
            </div>
            <div className="metric-card">
              <strong>Auth foundation</strong>
              <span>Login, register и password reset request подключены к backend API.</span>
            </div>
            <div className="metric-card">
              <strong>Без выдуманного API</strong>
              <span>UI Slice 1 строится только на существующих OpenAPI-ручках.</span>
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
              title="Не удалось загрузить превью каталога"
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
