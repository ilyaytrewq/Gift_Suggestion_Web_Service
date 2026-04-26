import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';

import {
  deleteCurrentWishlist,
  removeCurrentWishlistItem,
} from '../../../features/wishlist/api/wishlist';
import {
  currentWishlistQueryKey,
  removeWishlistGift,
  useWishlistQuery,
} from '../../../features/wishlist/model/use-wishlist';
import { useAuth } from '../../../shared/auth/use-auth';
import { formatPrice } from '../../../shared/lib/format';
import { Button } from '../../../shared/ui/button/button';
import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { Card } from '../../../shared/ui/card/card';
import { EmptyState } from '../../../shared/ui/feedback/empty-state';
import { ErrorBanner } from '../../../shared/ui/feedback/error-banner';
import { PageLoader } from '../../../shared/ui/feedback/page-loader';
import { Container } from '../../../shared/ui/layout/container';

const FALLBACK_IMAGE =
  'https://images.unsplash.com/photo-1513475382585-d06e58bcb0ff?auto=format&fit=crop&w=900&q=80';

export function WishlistPage(): JSX.Element {
  const auth = useAuth();
  const queryClient = useQueryClient();
  const wishlistQuery = useWishlistQuery();
  const userID = auth.user?.id ?? null;
  const queryKey = currentWishlistQueryKey(userID);

  const removeMutation = useMutation({
    mutationFn: removeCurrentWishlistItem,
    onSuccess: (_payload, giftID) => {
      queryClient.setQueryData(queryKey, (current: typeof wishlistQuery.data) => (
        removeWishlistGift(current, giftID)
      ));
    },
  });

  const clearMutation = useMutation({
    mutationFn: deleteCurrentWishlist,
    onSuccess: () => {
      queryClient.setQueryData(queryKey, (current: typeof wishlistQuery.data) => {
        if (!current) {
          return current;
        }

        return {
          ...current,
          data: {
            ...current.data,
            wishlist: {
              ...current.data.wishlist,
              item_count: 0,
              items: [],
            },
          },
        };
      });
    },
  });

  if (!auth.accessToken) {
    return (
      <Container className="page-stack">
        <EmptyState
          action={(
            <Link className={buttonClassName()} to="/login?next=%2Fwishlist">
              Войдите, чтобы сохранить подарок
            </Link>
          )}
          description="После входа здесь будут появляться сохранённые подарки."
          title="Список желаний"
        />
      </Container>
    );
  }

  if (wishlistQuery.isPending) {
    return (
      <PageLoader
        title="Загружаем список желаний"
        description="Собираем сохранённые подарки."
      />
    );
  }

  if (wishlistQuery.isError) {
    return (
      <Container className="page-stack">
        <ErrorBanner
          error={wishlistQuery.error}
          title="Не удалось открыть список желаний"
        />
      </Container>
    );
  }

  const wishlist = wishlistQuery.data?.data.wishlist;

  if (!wishlist || wishlist.items.length === 0) {
    return (
      <Container className="page-stack">
        <section className="section-heading">
          <p className="eyebrow">Список желаний</p>
          <h1>Список желаний</h1>
        </section>
        <EmptyState
          action={(
            <div className="wishlist-empty-actions">
              <Link className={buttonClassName()} to="/catalog">
                Перейти в каталог
              </Link>
              <Link className={buttonClassName({ variant: 'ghost' })} to="/recommendation">
                Подобрать подарок
              </Link>
            </div>
          )}
          description="В вашем списке пока нет подарков"
          title="Список желаний"
        />
      </Container>
    );
  }

  return (
    <Container className="page-stack">
      <section className="section-heading section-heading--inline">
        <div>
          <p className="eyebrow">Список желаний</p>
          <h1>Список желаний</h1>
          <p className="page-copy">
            Сохранено подарков: {wishlist.item_count}
          </p>
        </div>
        <div className="wishlist-toolbar">
          <Button
            disabled={clearMutation.isPending}
            variant="ghost"
            type="button"
            onClick={() => {
              clearMutation.reset();
              void clearMutation.mutateAsync();
            }}
          >
            {clearMutation.isPending ? 'Очищаем...' : 'Очистить список'}
          </Button>
        </div>
      </section>

      {removeMutation.isError ? (
        <ErrorBanner
          error={removeMutation.error}
          title="Не удалось обновить список желаний"
        />
      ) : null}

      {clearMutation.isError ? (
        <ErrorBanner
          error={clearMutation.error}
          title="Не удалось очистить список желаний"
        />
      ) : null}

      <div className="gift-grid">
        {wishlist.items.map((item) => (
          <Card className="gift-card" key={item.id}>
            <Link className="gift-card__image-link" to={`/catalog/${item.gift.id}`}>
              <img
                alt={item.gift.name}
                className="gift-card__image"
                src={item.gift.image ?? FALLBACK_IMAGE}
              />
            </Link>

            <div className="gift-card__content">
              <div className="gift-card__meta">
                {item.gift.category ? (
                  <span className="chip chip--muted">{item.gift.category.name}</span>
                ) : null}
                {item.gift.age_restriction ? (
                  <span className="chip chip--muted">{item.gift.age_restriction}+</span>
                ) : null}
              </div>

              <div className="gift-card__heading">
                <h3>{item.gift.name}</h3>
                <strong>{formatPrice(item.gift.price)}</strong>
              </div>

              <p>{item.gift.description}</p>

              <div className="gift-card__actions">
                <Link className={buttonClassName()} to={`/catalog/${item.gift.id}`}>
                  Подробнее
                </Link>
                <Button
                  disabled={removeMutation.isPending}
                  variant="ghost"
                  type="button"
                  onClick={() => {
                    removeMutation.reset();
                    void removeMutation.mutateAsync(item.gift.id);
                  }}
                >
                  Убрать
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>
    </Container>
  );
}
