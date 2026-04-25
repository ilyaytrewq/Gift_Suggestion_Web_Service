import { useQuery } from '@tanstack/react-query';
import { Link, useLocation, useParams } from 'react-router-dom';

import { getCatalogGift } from '../../../entities/gift/api/gifts';
import { formatPrice } from '../../../shared/lib/format';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { EmptyState } from '../../../shared/ui/feedback/empty-state';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { Container } from '../../../shared/ui/layout/container';
import { GiftDetailSkeleton } from '../../../shared/ui/skeleton/skeleton';

const FALLBACK_IMAGE =
  'https://images.unsplash.com/photo-1513475382585-d06e58bcb0ff?auto=format&fit=crop&w=1200&q=80';

export function GiftPage(): JSX.Element {
  const { giftId } = useParams();
  const location = useLocation();

  const giftQuery = useQuery({
    enabled: Boolean(giftId),
    queryFn: () => getCatalogGift(giftId ?? ''),
    queryKey: ['gift', giftId],
  });

  const backHref =
    typeof location.state === 'object' &&
    location.state &&
    'from' in location.state &&
    typeof location.state.from === 'string'
      ? location.state.from
      : '/catalog';

  return (
    <Container className="page-stack">
      <Link className="back-link" to={backHref}>
        ← Вернуться в каталог
      </Link>

      {giftQuery.isLoading ? <GiftDetailSkeleton /> : null}

      {giftQuery.isError ? (
        <ErrorBanner error={giftQuery.error} title="Не удалось загрузить карточку подарка" />
      ) : null}

      {giftQuery.data ? (
        <section className="gift-detail">
          <div className="gift-detail__media">
            <img
              alt={giftQuery.data.data.gift.name}
              className="gift-detail__image"
              src={giftQuery.data.data.gift.image ?? FALLBACK_IMAGE}
            />
          </div>

          <div className="gift-detail__content">
            <div className="gift-detail__chips">
              {giftQuery.data.data.gift.category ? (
                <span className="chip">
                  {giftQuery.data.data.gift.category.name}
                </span>
              ) : null}
              {giftQuery.data.data.gift.age_restriction ? (
                <span className="chip">
                  {giftQuery.data.data.gift.age_restriction}+
                </span>
              ) : null}
            </div>

            <h1>{giftQuery.data.data.gift.name}</h1>
            <p className="gift-detail__price">
              {formatPrice(giftQuery.data.data.gift.price)}
            </p>
            <p className="gift-detail__description">
              {giftQuery.data.data.gift.description}
            </p>

            <div className="gift-detail__actions">
              <a
                className={buttonClassName({ size: 'lg' })}
                href={giftQuery.data.data.gift.store_link}
                rel="noreferrer"
                target="_blank"
              >
                Перейти к покупке
              </a>
              <Link
                className={buttonClassName({ size: 'lg', variant: 'secondary' })}
                to="/catalog"
              >
                Смотреть другие идеи
              </Link>
            </div>

            <div className="gift-detail__note">
              <strong>Учтённый gap:</strong> backend сейчас отдаёт один `store_link`,
              поэтому карточка строится вокруг одного purchase CTA, а не списка магазинов.
            </div>
          </div>
        </section>
      ) : null}

      {!giftId ? (
        <EmptyState
          description="Откройте карточку из каталога, чтобы увидеть детали подарка."
          title="Идентификатор подарка не передан"
        />
      ) : null}
    </Container>
  );
}
