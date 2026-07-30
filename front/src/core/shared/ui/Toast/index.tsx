'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { $toasts, toastDismissed } from './model';
import styles from './Toast.module.css';

export const ToastViewport = (): JSX.Element | null => {
  const [toasts, dismiss] = useUnit([$toasts, toastDismissed]);

  useEffect(() => {
    if (toasts.length === 0) {
      return;
    }

    const timers = toasts.map((toast) =>
      window.setTimeout(() => {
        dismiss(toast.id);
      }, 3200),
    );

    return () => {
      timers.forEach((timer) => window.clearTimeout(timer));
    };
  }, [dismiss, toasts]);

  if (toasts.length === 0) {
    return null;
  }

  return (
    <div className={clsx(styles.viewport)}>
      {toasts.map((toast) => (
        <div
          className={clsx(styles.toast, toast.tone === 'error' && styles.toastError)}
          key={toast.id}
          role="status"
        >
          <span>{toast.message}</span>
          <button
            aria-label="Закрыть"
            className={clsx(styles.close)}
            onClick={() => dismiss(toast.id)}
            type="button"
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
};
