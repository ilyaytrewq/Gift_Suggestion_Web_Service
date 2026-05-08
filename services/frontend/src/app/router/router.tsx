import {
  BrowserRouter,
  Navigate,
  Route,
  Routes,
} from 'react-router-dom';

import { VkOAuthCallbackPage } from '../../pages/auth/ui/vk-oauth-callback-page';
import { EmailVerifyPage } from '../../pages/auth/ui/email-verify-page';
import { LoginPage } from '../../pages/auth/ui/login-page';
import { PasswordResetPage } from '../../pages/auth/ui/password-reset-page';
import { PasswordResetConfirmPage } from '../../pages/auth/ui/password-reset-confirm-page';
import { RegisterPage } from '../../pages/auth/ui/register-page';
import { CatalogPage } from '../../pages/catalog/ui/catalog-page';
import { GiftPage } from '../../pages/gift/ui/gift-page';
import { HomePage } from '../../pages/home/ui/home-page';
import { NotFoundPage } from '../../pages/not-found/ui/not-found-page';
import { ProfilePage } from '../../pages/profile/ui/profile-page';
import { RecommendationPage } from '../../pages/recommendation/ui/recommendation-page';
import { WishlistPage } from '../../pages/wishlist/ui/wishlist-page';
import { AdminImportPage } from '../../pages/admin/ui/admin-import-page';
import { AppShell } from '../../shared/ui/layout/app-shell';

export function AppRouter(): JSX.Element {
  return (
    <BrowserRouter>
      <AppShell>
        <Routes>
          <Route element={<HomePage />} path="/" />
          <Route element={<CatalogPage />} path="/catalog" />
          <Route element={<GiftPage />} path="/catalog/:giftId" />
          <Route element={<RecommendationPage />} path="/recommendation" />
          <Route element={<WishlistPage />} path="/wishlist" />
          <Route element={<ProfilePage />} path="/profile" />
          <Route element={<Navigate replace to="/admin/import" />} path="/admin" />
          <Route element={<AdminImportPage />} path="/admin/import" />
          <Route element={<LoginPage />} path="/login" />
          <Route element={<RegisterPage />} path="/register" />
          <Route element={<PasswordResetPage />} path="/password-reset" />
          <Route element={<PasswordResetConfirmPage />} path="/password-reset/confirm" />
          <Route element={<EmailVerifyPage />} path="/auth/email-verify" />
          <Route element={<VkOAuthCallbackPage />} path="/auth/vk-callback" />
          <Route element={<NotFoundPage />} path="*" />
        </Routes>
      </AppShell>
    </BrowserRouter>
  );
}
