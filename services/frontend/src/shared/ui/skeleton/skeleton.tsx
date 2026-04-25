import { Card } from '../card/card';

export function Skeleton({
  className,
}: {
  className?: string;
}): JSX.Element {
  return <span className={`skeleton ${className ?? ''}`.trim()} />;
}

export function CatalogGridSkeleton(): JSX.Element {
  return (
    <div className="gift-grid">
      {Array.from({ length: 8 }).map((_, index) => (
        <Card className="gift-card" key={index}>
          <Skeleton className="gift-card__image-skeleton" />
          <Skeleton className="skeleton--title" />
          <Skeleton className="skeleton--text" />
          <Skeleton className="skeleton--text" />
        </Card>
      ))}
    </div>
  );
}

export function GiftDetailSkeleton(): JSX.Element {
  return (
    <section className="gift-detail">
      <Skeleton className="gift-detail__media-skeleton" />
      <div className="gift-detail__content">
        <Skeleton className="skeleton--eyebrow" />
        <Skeleton className="skeleton--hero-title" />
        <Skeleton className="skeleton--text" />
        <Skeleton className="skeleton--text" />
        <Skeleton className="skeleton--button" />
      </div>
    </section>
  );
}
