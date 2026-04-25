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

export type RecommendationRequest = JsonBody<
  paths['/api/v1/recommendations']['post']
>;
export type RecommendationResponse = JsonResponse<
  paths['/api/v1/recommendations']['post']['responses']['200']
>;
