import type { components, paths } from './generated/schema';

type JsonBody<T> = T extends {
  requestBody?: { content: { 'application/json': infer U } };
}
  ? U
  : never;

type JsonResponse<T> = T extends {
  content: { 'application/json': infer U };
}
  ? U
  : never;

export type ErrorEnvelope = components['schemas']['ErrorEnvelope'];
export type Meta = components['schemas']['Meta'];
export type UserProfile = components['schemas']['UserProfile'];
export type CatalogGift = components['schemas']['CatalogGift'];
export type CatalogCategory = components['schemas']['CatalogCategory'];
export type Page = components['schemas']['Page'];

export type ListCatalogGiftsQuery = NonNullable<
  paths['/api/v1/catalog/gifts']['get']['parameters']['query']
>;
export type ListCatalogGiftsResponse = JsonResponse<
  paths['/api/v1/catalog/gifts']['get']['responses']['200']
>;

export type GetCatalogGiftResponse = JsonResponse<
  paths['/api/v1/catalog/gifts/{gift_id}']['get']['responses']['200']
>;

export type ListCatalogCategoriesQuery = NonNullable<
  paths['/api/v1/catalog/categories']['get']['parameters']['query']
>;
export type ListCatalogCategoriesResponse = JsonResponse<
  paths['/api/v1/catalog/categories']['get']['responses']['200']
>;

export type LoginRequest = JsonBody<paths['/api/v1/auth/login']['post']>;
export type LoginResponse = JsonResponse<
  paths['/api/v1/auth/login']['post']['responses']['200']
>;

export type RegisterRequest = JsonBody<paths['/api/v1/users']['post']>;
export type RegisterResponse = JsonResponse<
  paths['/api/v1/users']['post']['responses']['201']
>;

export type PasswordResetRequest = JsonBody<
  paths['/api/v1/auth/password-reset/request']['post']
>;
export type PasswordResetResponse = JsonResponse<
  paths['/api/v1/auth/password-reset/request']['post']['responses']['202']
>;

export type PasswordResetConfirmRequest = JsonBody<
  paths['/api/v1/auth/password-reset/confirm']['post']
>;
export type PasswordResetConfirmResponse = JsonResponse<
  paths['/api/v1/auth/password-reset/confirm']['post']['responses']['200']
>;

export type EmailVerificationConfirmRequest = JsonBody<
  paths['/api/v1/auth/email-verification/confirm']['post']
>;
export type EmailVerificationConfirmResponse = JsonResponse<
  paths['/api/v1/auth/email-verification/confirm']['post']['responses']['200']
>;

export type RefreshResponse = JsonResponse<
  paths['/api/v1/auth/refresh']['post']['responses']['200']
>;
export type LogoutResponse = JsonResponse<
  paths['/api/v1/auth/logout']['post']['responses']['200']
>;

export type CurrentUserResponse = JsonResponse<
  paths['/api/v1/users/me']['get']['responses']['200']
>;
export type UpdateCurrentUserRequest = JsonBody<
  paths['/api/v1/users/me']['patch']
>;
export type UpdateCurrentUserResponse = JsonResponse<
  paths['/api/v1/users/me']['patch']['responses']['200']
>;

export type PromoteUserToAdminRequest = JsonBody<
  paths['/api/v1/admin/users/promote']['post']
>;
export type PromoteUserToAdminResponse = JsonResponse<
  paths['/api/v1/admin/users/promote']['post']['responses']['200']
>;

export type RecommendationRequest = JsonBody<
  paths['/api/v1/recommendations']['post']
>;
export type RecommendationResponse = JsonResponse<
  paths['/api/v1/recommendations']['post']['responses']['200']
>;

export type Wishlist = components['schemas']['Wishlist'];
export type WishlistItem = components['schemas']['WishlistItem'];

export type GetCurrentWishlistResponse = JsonResponse<
  paths['/api/v1/wishlist']['get']['responses']['200']
>;
export type DeleteCurrentWishlistResponse = JsonResponse<
  paths['/api/v1/wishlist']['delete']['responses']['200']
>;
export type AddCurrentWishlistItemRequest = JsonBody<
  paths['/api/v1/wishlist/items']['post']
>;
export type AddCurrentWishlistItemResponse = JsonResponse<
  paths['/api/v1/wishlist/items']['post']['responses']['200']
>;
export type RemoveCurrentWishlistItemResponse = JsonResponse<
  paths['/api/v1/wishlist/items/{gift_id}']['delete']['responses']['200']
>;

export type GetSimilarGiftsResponse = JsonResponse<
  paths['/api/v1/catalog/gifts/{gift_id}/similar']['get']['responses']['200']
>;

export type ImportJob = components['schemas']['ImportJob'];
export type CreateImportJobResponse = JsonResponse<
  paths['/api/v1/admin/import-jobs']['post']['responses']['201']
>;
export type GetImportJobResponse = JsonResponse<
  paths['/api/v1/admin/import-jobs/{job_id}']['get']['responses']['200']
>;
export type GetImportJobErrorsResponse = JsonResponse<
  paths['/api/v1/admin/import-jobs/{job_id}/errors']['get']['responses']['200']
>;

export type TrackEventRequest = JsonBody<
  paths['/api/v1/tracking/events']['post']
>;
export type TrackEventResponse = JsonResponse<
  paths['/api/v1/tracking/events']['post']['responses']['200']
>;

export type VKConnectionEnvelope = JsonResponse<
  paths['/api/v1/integrations/vk/connection']['get']['responses']['200']
>;
export type ConnectVKRequestBody = JsonBody<
  paths['/api/v1/integrations/vk/connection']['put']
>;
