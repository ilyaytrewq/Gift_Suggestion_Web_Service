import {
  BrowserRouter,
  Route,
  Routes,
} from 'react-router-dom';

import { LoginPage } from '../../pages/auth/ui/login-page';
import { PasswordResetPage } from '../../pages/auth/ui/password-reset-page';
import { RegisterPage } from '../../pages/auth/ui/register-page';
import { CatalogPage } from '../../pages/catalog/ui/catalog-page';
import { GiftPage } from '../../pages/gift/ui/gift-page';
import { HomePage } from '../../pages/home/ui/home-page';
import { NotFoundPage } from '../../pages/not-found/ui/not-found-page';
import { AppShell } from '../../shared/ui/layout/app-shell';

export function AppRouter(): JSX.Element {
  return (
    <BrowserRouter>
      <AppShell>
        <Routes>
          <Route element={<HomePage />} path="/" />
          <Route element={<CatalogPage />} path="/catalog" />
          <Route element={<GiftPage />} path="/catalog/:giftId" />
          <Route element={<LoginPage />} path="/login" />
          <Route element={<RegisterPage />} path="/register" />
          <Route element={<PasswordResetPage />} path="/password-reset" />
          <Route element={<NotFoundPage />} path="*" />
        </Routes>
      </AppShell>
    </BrowserRouter>
  );
}
