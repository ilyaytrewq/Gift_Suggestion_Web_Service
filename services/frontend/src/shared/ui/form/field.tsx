import type { PropsWithChildren, ReactNode } from 'react';

import { cn } from '../../lib/cn';

export function Field({
  children,
  error,
  hint,
  label,
}: PropsWithChildren<{
  error?: string;
  hint?: ReactNode;
  label: string;
}>): JSX.Element {
  return (
    <label className="field">
      <span className="field__label">{label}</span>
      {children}
      {hint ? <span className="field__hint">{hint}</span> : null}
      {error ? <span className={cn('field__error')}>{error}</span> : null}
    </label>
  );
}
