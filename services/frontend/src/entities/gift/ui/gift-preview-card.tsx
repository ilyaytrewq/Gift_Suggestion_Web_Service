import { Link, useLocation } from 'react-router-dom';

import type { CatalogGift } from '../../../shared/api/contracts';
import { formatPrice } from '../../../shared/lib/format';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { Card } from '../../../shared/ui/card/card';

const FALLBACK_IMAGE =
  'https://images.unsplash.com/photo-1513475382585-d06e58bcb0ff?auto=format&fit=crop&w=900&q=80';

export function GiftPreviewCard({
  gift,
}: {
  gift: CatalogGift;
}): JSX.Element {
  const location = useLocation();

  return (
    <Card className="gift-card">
      <Link
        className="gift-card__image-link"
        to={`/catalog/${gift.id}`}
        state={{ from: location.pathname + location.search }}
      >
        <img
          alt={gift.name}
          className="gift-card__image"
          src={gift.image ?? FALLBACK_IMAGE}
        />
      </Link>

      <div className="gift-card__content">
        <div className="gift-card__meta">
          {gift.category ? <span className="chip chip--muted">{gift.category.name}</span> : null}
          {gift.age_restriction ? (
            <span className="chip chip--muted">{gift.age_restriction}+</span>
          ) : null}
        </div>

        <div className="gift-card__heading">
          <h3>{gift.name}</h3>
          <strong>{formatPrice(gift.price)}</strong>
        </div>

        <p>{gift.description}</p>

        <div className="gift-card__actions">
          <Link className={buttonClassName()} to={`/catalog/${gift.id}`} state={{ from: location.pathname + location.search }}>
            Открыть карточку
          </Link>
          <a
            className={buttonClassName({ variant: 'ghost' })}
            href={gift.store_link}
            rel="noreferrer"
            target="_blank"
          >
            Купить
          </a>
        </div>
      </div>
    </Card>
  );
}
