import type {
  AddCurrentWishlistItemRequest,
  AddCurrentWishlistItemResponse,
  DeleteCurrentWishlistResponse,
  GetCurrentWishlistResponse,
  RemoveCurrentWishlistItemResponse,
} from '../../../shared/api/contracts';
import { requestJson } from '../../../shared/api/http';

export function getCurrentWishlist(): Promise<GetCurrentWishlistResponse> {
  return requestJson<GetCurrentWishlistResponse>('/api/v1/wishlist', {
    auth: true,
  });
}

export function addCurrentWishlistItem(
  payload: AddCurrentWishlistItemRequest,
): Promise<AddCurrentWishlistItemResponse> {
  return requestJson<AddCurrentWishlistItemResponse>('/api/v1/wishlist/items', {
    method: 'POST',
    auth: true,
    body: payload,
  });
}

export function removeCurrentWishlistItem(
  giftID: string,
): Promise<RemoveCurrentWishlistItemResponse> {
  return requestJson<RemoveCurrentWishlistItemResponse>(`/api/v1/wishlist/items/${giftID}`, {
    method: 'DELETE',
    auth: true,
  });
}

export function deleteCurrentWishlist(): Promise<DeleteCurrentWishlistResponse> {
  return requestJson<DeleteCurrentWishlistResponse>('/api/v1/wishlist', {
    method: 'DELETE',
    auth: true,
  });
}
