import type { ReactNode } from 'react';

export function EmptyState({
  action,
  description,
  title,
}: {
  action?: ReactNode;
  description: string;
  title: string;
}): JSX.Element {
  return (
    <div className="empty-state">
      <p className="eyebrow">Пусто</p>
      <h2>{title}</h2>
      <p>{description}</p>
      {action ? <div className="empty-state__action">{action}</div> : null}
    </div>
  );
}
