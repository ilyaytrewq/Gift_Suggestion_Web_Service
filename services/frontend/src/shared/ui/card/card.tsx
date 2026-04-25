import type { HTMLAttributes, PropsWithChildren } from 'react';

import { cn } from '../../lib/cn';

export function Card({
  children,
  className,
  ...props
}: PropsWithChildren<HTMLAttributes<HTMLDivElement>>): JSX.Element {
  return (
    <div className={cn('card', className)} {...props}>
      {children}
    </div>
  );
}
