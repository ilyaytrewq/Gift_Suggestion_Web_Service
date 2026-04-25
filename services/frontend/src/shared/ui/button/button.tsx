import type { ButtonHTMLAttributes } from 'react';
import type { ButtonStyleOptions } from './button-class-name';
import { buttonClassName } from './button-class-name';

export function Button({
  children,
  className,
  size,
  variant,
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & {
  size?: ButtonStyleOptions['size'];
  variant?: ButtonStyleOptions['variant'];
}): JSX.Element {
  return (
    <button
      className={[buttonClassName({ size, variant }), className]
        .filter(Boolean)
        .join(' ')}
      {...props}
    >
      {children}
    </button>
  );
}
