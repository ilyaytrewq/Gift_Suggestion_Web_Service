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
import { ApiError, getUserFacingApiErrorMessage } from '../../../shared/api/api-error';
import { useAuth } from '../../../shared/auth/use-auth';
import { cn } from '../../../shared/lib/cn';
import { buttonClassName, type ButtonStyleOptions } from '../../../shared/ui/button/button-class-name';
import { useToast } from '../../../shared/ui/toast/use-toast';

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
  const toast = useToast();

  const mutation = useMutation({
    mutationFn: () => addCurrentWishlistItem({ gift_id: giftID }),
    onSuccess: (payload) => {
      queryClient.setQueryData(queryKey, (current: typeof wishlistQuery.data) => {
        if (!current) {
          void queryClient.invalidateQueries({ queryKey });
          return current;
        }
        return appendWishlistItem(current, payload.data.item) ?? current;
      });
      track({ type: 'wishlist_add', gift_id: giftID });
      toast.show({
        variant: 'success',
        message: payload.data.already_in_wishlist
          ? 'Этот подарок уже в вашем списке желаний.'
          : 'Подарок сохранён в список желаний.',
      });
    },
    onError: (error) => {
      const message =
        error instanceof ApiError
          ? getUserFacingApiErrorMessage(error)
          : 'Не удалось сохранить подарок. Попробуйте ещё раз.';
      toast.show({ variant: 'error', message });
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
      <span className={cn('wishlist-save-button--saved', className)} title="Товар уже в списке желаний">
        <button className={cn(buttonClassName({ size, variant }), 'wishlist-save-button--saved-btn')} disabled type="button">
          Сохранено
        </button>
      </span>
    );
  }

  return (
    <div className={cn('wishlist-action', className)}>
      <button
        aria-busy={isSaving}
        className={buttonClassName({ size, variant })}
        disabled={isSaving}
        type="button"
        onClick={() => mutation.mutate()}
      >
        {isSaving ? 'Сохраняем...' : 'Сохранить'}
      </button>
    </div>
  );
}
