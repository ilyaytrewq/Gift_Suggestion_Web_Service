import type { PropsWithChildren } from 'react';

import { cn } from '../../lib/cn';

export function Notice({
  children,
  tone = 'info',
}: PropsWithChildren<{ tone?: 'info' | 'success' }>): JSX.Element {
  return <div className={cn('banner', `banner--${tone}`)}>{children}</div>;
}
