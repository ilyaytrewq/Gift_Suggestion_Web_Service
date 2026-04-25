import type { PropsWithChildren } from 'react';

import { cn } from '../../lib/cn';

export function Container({
  children,
  className,
}: PropsWithChildren<{ className?: string }>): JSX.Element {
  return <div className={cn('container', className)}>{children}</div>;
}
