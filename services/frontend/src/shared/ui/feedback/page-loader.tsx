export function PageLoader({
  description,
  title,
}: {
  description: string;
  title: string;
}): JSX.Element {
  return (
    <div className="page-loader">
      <span className="page-loader__spinner" aria-hidden="true" />
      <div>
        <h2>{title}</h2>
        <p>{description}</p>
      </div>
    </div>
  );
}
