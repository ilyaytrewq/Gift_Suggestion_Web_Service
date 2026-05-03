import { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';

import { logoutUser } from '../../../features/auth/api/auth';
import { useAuth } from '../../auth/use-auth';
import { buttonClassName } from '../button/button-class-name';
import { Container } from './container';

export function Header(): JSX.Element {
  const auth = useAuth();
  const navigate = useNavigate();
  const [isLoggingOut, setIsLoggingOut] = useState(false);

  async function handleLogout(): Promise<void> {
    if (isLoggingOut) {
      return;
    }

    setIsLoggingOut(true);
    try {
      await logoutUser();
    } catch {
      // Logout remains idempotent on UI side even when backend session is already gone.
    } finally {
      auth.clearSession();
      setIsLoggingOut(false);
      navigate('/', { replace: true });
    }
  }

  return (
    <header className="site-header">
      <Container className="site-header__inner">
        <Link className="site-header__brand" to="/">
          Gift Suggestion
        </Link>

        <nav className="site-header__nav" aria-label="Главная навигация">
          <Link className="site-header__link" to="/catalog">
            Каталог
          </Link>
          {auth.user ? (
            <div className="site-header__actions">
              <Link className="site-header__link" to="/wishlist">
                Список желаний
              </Link>
              <span className="site-header__user">
                {auth.user.display_name || auth.user.email}
              </span>
              <button
                className={buttonClassName({ variant: 'ghost' })}
                type="button"
                onClick={handleLogout}
                disabled={isLoggingOut}
              >
                {isLoggingOut ? 'Выход...' : 'Выйти'}
              </button>
            </div>
          ) : (
            <div className="site-header__actions">
              <Link className="site-header__link" to="/login">
                Войти
              </Link>
              <Link className={buttonClassName({ variant: 'secondary' })} to="/register">
                Создать аккаунт
              </Link>
            </div>
          )}
        </nav>
      </Container>
    </header>
  );
}
