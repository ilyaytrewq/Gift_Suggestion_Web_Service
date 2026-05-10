import { useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link, useLocation, useParams } from 'react-router-dom';

import { getCatalogGift, getSimilarGifts } from '../../../entities/gift/api/gifts';
import { useTrackEvent } from '../../../features/tracking/model/use-track-event';
import { WishlistSaveButton } from '../../../features/wishlist/ui/wishlist-save-button';
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
  const track = useTrackEvent();

  const giftQuery = useQuery({
    enabled: Boolean(giftId),
    queryFn: () => getCatalogGift(giftId ?? ''),
    queryKey: ['gift', giftId],
  });

  const similarQuery = useQuery({
    enabled: Boolean(giftId),
    queryFn: () => getSimilarGifts(giftId ?? ''),
    queryKey: ['gift-similar', giftId],
  });

  const backHref =
    typeof location.state === 'object' &&
    location.state &&
    'from' in location.state &&
    typeof location.state.from === 'string'
      ? location.state.from
      : '/catalog';

  const gift = giftQuery.data?.data.gift;
  const offers = gift?.offers ?? [];
  const similarGifts = similarQuery.data?.data.items ?? [];

  useEffect(() => {
    if (!gift) return;
    track({
      type: 'card_view',
      gift_id: gift.id,
      metadata: { surface: 'direct' },
    });
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [giftQuery.dataUpdatedAt]);

  return (
    <Container className="page-stack">
      <Link className="back-link" to={backHref}>
        ← Вернуться в каталог
      </Link>

      {giftQuery.isLoading ? <GiftDetailSkeleton /> : null}

      {giftQuery.isError ? (
        <ErrorBanner error={giftQuery.error} title="Не удалось открыть подарок" />
      ) : null}

      {gift ? (
        <>
          <section className="gift-detail">
            <div className="gift-detail__media">
              <img
                alt={gift.name}
                className="gift-detail__image"
                src={gift.image ?? FALLBACK_IMAGE}
              />
            </div>

            <div className="gift-detail__content">
              <div className="gift-detail__chips">
                {gift.category ? (
                  <span className="chip">{gift.category.name}</span>
                ) : null}
                {gift.age_restriction ? (
                  <span className="chip">{gift.age_restriction}+</span>
                ) : null}
              </div>

              <h1>{gift.name}</h1>
              <p className="gift-detail__price">{formatPrice(gift.price)}</p>
              <p className="gift-detail__description">{gift.description}</p>

              {/* Shops / offers list */}
              {offers.length > 0 ? (
                <div>
                  <h2 style={{ fontSize: '1rem', fontWeight: 600, margin: '1rem 0 0.5rem' }}>
                    Где купить
                  </h2>
                  <div className="gift-offers">
                    {offers.map((offer) => (
                      <div
                        key={offer.id}
                        className={[
                          'gift-offer',
                          offer.available ? '' : 'gift-offer--unavailable',
                        ].join(' ')}
                      >
                        <span className="gift-offer__name">{offer.store_name}</span>
                        <span className="gift-offer__price">
                          {formatPrice(offer.price)} {offer.currency}
                        </span>
                        {offer.available ? (
                          <a
                            className={buttonClassName({ variant: 'secondary' })}
                            href={offer.store_url}
                            rel="noreferrer"
                            target="_blank"
                            onClick={() =>
                              track({
                                type: 'outbound_click',
                                gift_id: gift.id,
                                metadata: { surface: 'direct' },
                              })
                            }
                          >
                            Купить
                          </a>
                        ) : (
                          <span style={{ fontSize: '0.8rem', color: 'var(--color-muted)' }}>
                            Нет в наличии
                          </span>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ) : null}

              <div className="gift-detail__actions">
                <a
                  className={buttonClassName({ size: 'lg' })}
                  href={gift.store_link}
                  rel="noreferrer"
                  target="_blank"
                  onClick={() =>
                    track({
                      type: 'outbound_click',
                      gift_id: gift.id,
                      metadata: { surface: 'direct' },
                    })
                  }
                >
                  В магазин
                </a>
                <WishlistSaveButton giftID={gift.id} size="lg" />
                <Link
                  className={buttonClassName({ size: 'lg', variant: 'secondary' })}
                  to="/catalog"
                >
                  Смотреть другие идеи
                </Link>
              </div>
            </div>
          </section>

          {/* Similar gifts / alternatives */}
          {similarGifts.length > 0 && (
            <section className="page-section">
              <h2 className="section-heading__title" style={{ fontSize: '1.2rem', marginBottom: '1rem' }}>
                Альтернативы
              </h2>
              <div className="gift-grid">
                {similarGifts.map((similar) => (
                  <article className="card gift-card" key={similar.id}>
                    {similar.image && (
                      <Link className="gift-card__image-link" to={`/catalog/${similar.id}`}>
                        <img
                          alt={similar.name}
                          className="gift-card__image"
                          src={similar.image}
                        />
                      </Link>
                    )}
                    <div className="gift-card__content">
                      <div className="gift-card__heading">
                        <h3>{similar.name}</h3>
                        <strong>{formatPrice(similar.price)}</strong>
                      </div>
                      <p>{similar.description}</p>
                      <div className="gift-card__actions">
                        <Link
                          className={buttonClassName()}
                          state={{ from: location.pathname }}
                          to={`/catalog/${similar.id}`}
                        >
                          Подробнее
                        </Link>
                        <WishlistSaveButton giftID={similar.id} />
                      </div>
                    </div>
                  </article>
                ))}
              </div>
            </section>
          )}
        </>
      ) : null}

      {!giftId ? (
        <EmptyState
          description="Откройте карточку из каталога, чтобы увидеть детали подарка."
          title="Подарок не найден"
        />
      ) : null}
    </Container>
  );
}
