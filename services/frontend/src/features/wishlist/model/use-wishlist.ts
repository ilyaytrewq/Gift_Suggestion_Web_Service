import { useQuery } from '@tanstack/react-query';

import type {
  GetCurrentWishlistResponse,
  Wishlist,
} from '../../../shared/api/contracts';
import { useAuth } from '../../../shared/auth/use-auth';
import { getCurrentWishlist } from '../api/wishlist';

export function currentWishlistQueryKey(userID: string | null): readonly [string, string, string] {
  return ['wishlist', 'current', userID ?? 'guest'];
}

export function useWishlistQuery() {
  const auth = useAuth();
  const userID = auth.user?.id ?? null;

  return useQuery({
    queryKey: currentWishlistQueryKey(userID),
    queryFn: getCurrentWishlist,
    enabled: Boolean(auth.accessToken && userID),
  });
}

export function isGiftSavedInWishlist(wishlist: Wishlist | null | undefined, giftID: string): boolean {
  if (!wishlist) {
    return false;
  }

  return wishlist.items.some((item) => item.gift.id === giftID);
}

export function appendWishlistItem(
  current: GetCurrentWishlistResponse | undefined,
  nextItem: GetCurrentWishlistResponse['data']['wishlist']['items'][number],
): GetCurrentWishlistResponse | undefined {
  if (!current) {
    return current;
  }

  if (current.data.wishlist.items.some((item) => item.gift.id === nextItem.gift.id)) {
    return current;
  }

  return {
    ...current,
    data: {
      ...current.data,
      wishlist: {
        ...current.data.wishlist,
        item_count: current.data.wishlist.item_count + 1,
        items: [nextItem, ...current.data.wishlist.items],
      },
    },
  };
}

export function removeWishlistGift(
  current: GetCurrentWishlistResponse | undefined,
  giftID: string,
): GetCurrentWishlistResponse | undefined {
  if (!current) {
    return current;
  }

  const nextItems = current.data.wishlist.items.filter((item) => item.gift.id !== giftID);
  if (nextItems.length === current.data.wishlist.items.length) {
    return current;
  }

  return {
    ...current,
    data: {
      ...current.data,
      wishlist: {
        ...current.data.wishlist,
        item_count: nextItems.length,
        items: nextItems,
      },
    },
  };
}
