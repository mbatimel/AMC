import { createEvent } from 'effector';

export type ToastTone = 'error' | 'success';

export const toastShown = createEvent<{ message: string; tone?: ToastTone }>();
