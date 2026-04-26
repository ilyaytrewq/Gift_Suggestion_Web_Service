import { Link } from 'react-router-dom';

import { Container } from './container';

export function Footer(): JSX.Element {
  return (
    <footer className="site-footer">
      <Container className="site-footer__inner">
        <div>
          <strong>Gift Suggestion</strong>
          <p>Сервис, который помогает быстрее находить удачные идеи для подарка.</p>
        </div>
        <div className="site-footer__links">
          <Link to="/catalog">Каталог идей</Link>
          <Link to="/login">Войти</Link>
          <Link to="/register">Создать аккаунт</Link>
        </div>
      </Container>
    </footer>
  );
}
