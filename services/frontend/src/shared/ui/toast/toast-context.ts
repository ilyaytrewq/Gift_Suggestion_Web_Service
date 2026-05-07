import { createContext } from 'react';

export type ToastVariant = 'success' | 'error';

export type ToastShowParams = {
  message: string;
  variant: ToastVariant;
};

export type ToastContextValue = {
  show: (toast: ToastShowParams) => void;
};

export const ToastContext = createContext<ToastContextValue | null>(null);
