'use client';

import { Toast, toast } from '@heroui/react';
import { createEffect, sample } from 'effector';

import { toastShown, type ToastTone } from './model';

const presentToastFx = createEffect(
  ({ message, tone }: { message: string; tone?: ToastTone }): void => {
    if (tone === 'error') {
      toast.danger(message);

      return;
    }

    toast.success(message);
  },
);

sample({
  clock: toastShown,
  target: presentToastFx,
});

/** Провайдер HeroUI Toast; показ идёт через `toastShown` → `presentToastFx`. */
export const ToastViewport = (): JSX.Element => {
  return <Toast.Provider placement="bottom end" />;
};
