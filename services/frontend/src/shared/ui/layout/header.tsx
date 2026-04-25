import { Link } from 'react-router-dom';

import { useAuth } from '../../auth/use-auth';
import { buttonClassName } from '../button/button-class-name';
import { Container } from './container';

export function Header(): JSX.Element {
  const auth = useAuth();

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
          <a className="site-header__link" href="#how-it-works">
            Как это работает
          </a>
          {auth.user ? (
            <span className="site-header__user">
              {auth.user.display_name || auth.user.email}
            </span>
          ) : (
            <div className="site-header__actions">
              <Link className="site-header__link" to="/login">
                Войти
              </Link>
              <Link className={buttonClassName({ variant: 'secondary' })} to="/register">
                Регистрация
              </Link>
            </div>
          )}
        </nav>
      </Container>
    </header>
  );
}
