import { createEffect, createEvent, createStore, sample } from 'effector';

import type { ContentPageKey, ContentPages } from '@/core/shared/api/content';

import { contentInvalidated } from '@/core/entities/content';
import { updateContentPageRequest } from '@/core/shared/api/content';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { toastShown } from '@/core/shared/ui/Toast/model';

export type ContentSavePayload = {
  key: ContentPageKey;
  value: ContentPages[ContentPageKey];
};

export const contentSaveRequested = createEvent<ContentSavePayload>();

export const saveContentFx = createEffect(async ({ key, value }: ContentSavePayload) =>
  updateContentPageRequest(key, value),
);

export const $isContentSaving = saveContentFx.pending;

export const $contentSaveError = createStore<null | string>(null)
  .on(saveContentFx, () => null)
  .on(saveContentFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось сохранить изменения'),
  );

sample({
  clock: contentSaveRequested,
  target: saveContentFx,
});

sample({
  clock: saveContentFx.done,
  fn: () => ({ message: 'Изменения сохранены', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: saveContentFx.done,
  target: contentInvalidated,
});
