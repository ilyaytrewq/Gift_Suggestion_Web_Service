import { useMutation, useQueryClient } from '@tanstack/react-query';
import { Link, useLocation } from 'react-router-dom';

import { addCurrentWishlistItem } from '../api/wishlist';
import {
  appendWishlistItem,
  currentWishlistQueryKey,
  isGiftSavedInWishlist,
  useWishlistQuery,
} from '../model/use-wishlist';
import { useTrackEvent } from '../../tracking/model/use-track-event';
import { useAuth } from '../../../shared/auth/use-auth';
import { cn } from '../../../shared/lib/cn';
import { buttonClassName, type ButtonStyleOptions } from '../../../shared/ui/button/button-class-name';

function buildLoginHref(pathname: string, search: string, hash: string): string {
  const nextPath = `${pathname}${search}${hash}`;
  return `/login?next=${encodeURIComponent(nextPath || '/')}`;
}

export function WishlistSaveButton({
  className,
  giftID,
  size = 'md',
  variant = 'secondary',
}: {
  className?: string;
  giftID: string;
  size?: ButtonStyleOptions['size'];
  variant?: ButtonStyleOptions['variant'];
}): JSX.Element {
  const auth = useAuth();
  const location = useLocation();
  const queryClient = useQueryClient();
  const wishlistQuery = useWishlistQuery();
  const userID = auth.user?.id ?? null;
  const queryKey = currentWishlistQueryKey(userID);
  const wishlist = wishlistQuery.data?.data.wishlist;
  const track = useTrackEvent();

  const mutation = useMutation({
    mutationFn: () => addCurrentWishlistItem({ gift_id: giftID }),
    onSuccess: (payload) => {
      queryClient.setQueryData(queryKey, (current: typeof wishlistQuery.data) => (
        appendWishlistItem(current, payload.data.item)
      ));
      track({ type: 'wishlist_add', gift_id: giftID });
    },
  });
  const isSaving = mutation.isPending;
  const hasSavedCurrentGift = mutation.data?.data.item.gift.id === giftID;
  const isSaved = isGiftSavedInWishlist(wishlist, giftID) || hasSavedCurrentGift;

  if (!auth.accessToken) {
    return (
      <Link
        className={cn(buttonClassName({ size, variant: 'ghost' }), className)}
        to={buildLoginHref(location.pathname, location.search, location.hash)}
      >
        Войдите, чтобы сохранить подарок
      </Link>
    );
  }

  if (isSaved) {
    return (
      <button
        className={cn(buttonClassName({ size, variant }), className)}
        disabled
        type="button"
      >
        Сохранено
      </button>
    );
  }

  return (
    <div className={cn('wishlist-action', className)}>
      <button
        aria-busy={isSaving}
        className={buttonClassName({ size, variant })}
        disabled={isSaving}
        type="button"
        onClick={() => {
          mutation.reset();
          void mutation.mutateAsync();
        }}
      >
        {isSaving ? 'Сохраняем...' : 'Сохранить'}
      </button>
      {mutation.isError ? (
        <p className="wishlist-action__hint">Не удалось сохранить подарок. Попробуйте ещё раз.</p>
      ) : null}
    </div>
  );
}
