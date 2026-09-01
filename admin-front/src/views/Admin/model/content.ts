import { createEffect, createEvent, createStore, sample } from 'effector';

import type { BannersSettings, ContentPageKey, ContentPages } from '@/core/shared/api/content';

import { contentInvalidated } from '@/core/entities/content';
import { updateBannersRequest, updateContentPageRequest } from '@/core/shared/api/content';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { toastShown } from '@/core/shared/ui/Toast/model';

export type ContentSavePayload = {
  key: ContentPageKey;
  value: ContentPages[ContentPageKey];
};

export const contentSaveRequested = createEvent<ContentSavePayload>();
export const bannersSaveRequested = createEvent<BannersSettings>();

export const saveContentFx = createEffect(async ({ key, value }: ContentSavePayload) =>
  updateContentPageRequest(key, value),
);

export const saveBannersFx = createEffect(async (payload: BannersSettings) =>
  updateBannersRequest(payload),
);

export const $isContentSaving = saveContentFx.pending;
export const $isBannersSaving = saveBannersFx.pending;

export const $contentSaveError = createStore<null | string>(null)
  .on([saveContentFx, saveBannersFx], () => null)
  .on([saveContentFx.failData, saveBannersFx.failData], (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось сохранить изменения'),
  );

sample({
  clock: contentSaveRequested,
  target: saveContentFx,
});

sample({
  clock: bannersSaveRequested,
  target: saveBannersFx,
});

sample({
  clock: [saveContentFx.done, saveBannersFx.done],
  fn: () => ({ message: 'Изменения сохранены', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: [saveContentFx.done, saveBannersFx.done],
  target: contentInvalidated,
});
