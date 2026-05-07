import {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type JSX,
  type PropsWithChildren,
} from 'react';

import {
  ToastContext,
  type ToastShowParams,
  type ToastVariant,
} from './toast-context';

type ToastEntry = ToastShowParams & { id: string };

const DEFAULT_DURATION_MS = 4200;

export function ToastProvider({ children }: PropsWithChildren): JSX.Element {
  const [toasts, setToasts] = useState<ToastEntry[]>([]);
  const timeoutsRef = useRef<Map<string, ReturnType<typeof setTimeout>>>(new Map());

  const dismiss = useCallback((id: string) => {
    const handle = timeoutsRef.current.get(id);
    if (handle !== undefined) {
      clearTimeout(handle);
      timeoutsRef.current.delete(id);
    }
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  useEffect(() => {
    const timers = timeoutsRef.current;
    return () => {
      timers.forEach((handle) => clearTimeout(handle));
      timers.clear();
    };
  }, []);

  const show = useCallback(
    (toast: ToastShowParams) => {
      const id =
        typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function'
          ? crypto.randomUUID()
          : `${Date.now()}-${Math.random().toString(16).slice(2)}`;

      setToasts((prev) => [...prev, { ...toast, id }]);

      const handle = setTimeout(() => dismiss(id), DEFAULT_DURATION_MS);
      timeoutsRef.current.set(id, handle);
    },
    [dismiss],
  );

  const value = useMemo(() => ({ show }), [show]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div aria-live="polite" className="toast-viewport" role="region">
        {toasts.map((t) => (
          <ToastItem key={t.id} message={t.message} variant={t.variant} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

function ToastItem({
  message,
  variant,
}: {
  message: string;
  variant: ToastVariant;
}): JSX.Element {
  const className = variant === 'success' ? 'toast toast--success' : 'toast toast--error';

  return <div className={className}>{message}</div>;
}
