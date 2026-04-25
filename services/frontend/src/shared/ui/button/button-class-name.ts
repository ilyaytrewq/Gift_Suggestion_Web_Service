import { cn } from '../../lib/cn';

type ButtonVariant = 'primary' | 'secondary' | 'ghost';
type ButtonSize = 'md' | 'lg';

export type ButtonStyleOptions = {
  block?: boolean;
  size?: ButtonSize;
  variant?: ButtonVariant;
};

export function buttonClassName(options?: ButtonStyleOptions): string {
  const variant = options?.variant ?? 'primary';
  const size = options?.size ?? 'md';

  return cn(
    'button',
    `button--${variant}`,
    `button--${size}`,
    options?.block && 'button--block',
  );
}
