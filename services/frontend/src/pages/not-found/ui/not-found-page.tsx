import { Link } from 'react-router-dom';

import { buttonClassName } from '../../../shared/ui/button/button-class-name';
import { EmptyState } from '../../../shared/ui/feedback/empty-state';
import { Container } from '../../../shared/ui/layout/container';

export function NotFoundPage(): JSX.Element {
  return (
    <Container className="page-stack">
      <EmptyState
        action={
          <Link className={buttonClassName()} to="/">
            На главную
          </Link>
        }
        description="Проверьте адрес или вернитесь в каталог идей."
        title="Страница не найдена"
      />
    </Container>
  );
}
