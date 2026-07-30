import { createEvent, createStore } from 'effector';

export type ToastItem = {
  id: string;
  message: string;
  tone?: 'error' | 'success';
};

export const toastShown = createEvent<{ message: string; tone?: ToastItem['tone'] }>();
export const toastDismissed = createEvent<string>();

export const $toasts = createStore<ToastItem[]>([])
  .on(toastShown, (state, payload) => [
    ...state,
    {
      id: `${Date.now()}-${Math.random().toString(16).slice(2)}`,
      message: payload.message,
      tone: payload.tone ?? 'success',
    },
  ])
  .on(toastDismissed, (state, id) => state.filter((toast) => toast.id !== id));
