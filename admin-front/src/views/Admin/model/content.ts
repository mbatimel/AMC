import { createEffect, createEvent, createStore, sample } from 'effector';

import type {
  BannersSettings,
  ContentPageKey,
  ContentPages,
  LegalDoc,
} from '@/core/shared/api/content';

import { contentInvalidated } from '@/core/entities/content';
import {
  updateBannersRequest,
  updateContentPageRequest,
  updateLegalDocRequest,
} from '@/core/shared/api/content';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';
import { toastShown } from '@/core/shared/ui/Toast/model';

export type ContentSavePayload = {
  key: ContentPageKey;
  value: ContentPages[ContentPageKey];
};

export type LegalSavePayload = {
  body: string;
  docId: string;
  summary: string;
  version: string;
};

export const contentSaveRequested = createEvent<ContentSavePayload>();
export const bannersSaveRequested = createEvent<BannersSettings>();
export const legalSaveRequested = createEvent<LegalSavePayload>();

export const saveContentFx = createEffect(async ({ key, value }: ContentSavePayload) =>
  updateContentPageRequest(key, value),
);

export const saveBannersFx = createEffect(async (payload: BannersSettings) =>
  updateBannersRequest(payload),
);

export const saveLegalFx = createEffect(
  async ({ body, docId, summary, version }: LegalSavePayload): Promise<LegalDoc> =>
    updateLegalDocRequest(docId, { body, summary, version }),
);

export const $isContentSaving = saveContentFx.pending;
export const $isBannersSaving = saveBannersFx.pending;
export const $isLegalSaving = saveLegalFx.pending;

export const $contentSaveError = createStore<null | string>(null)
  .on([saveContentFx, saveBannersFx, saveLegalFx], () => null)
  .on([saveContentFx.failData, saveBannersFx.failData, saveLegalFx.failData], (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось сохранить изменения'),
  );

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> fn -> target */

sample({
  clock: contentSaveRequested,
  target: saveContentFx,
});

sample({
  clock: bannersSaveRequested,
  target: saveBannersFx,
});

sample({
  clock: legalSaveRequested,
  target: saveLegalFx,
});

sample({
  clock: [saveContentFx.done, saveBannersFx.done, saveLegalFx.done],
  fn: () => ({ message: 'Изменения сохранены', tone: 'success' as const }),
  target: toastShown,
});

sample({
  clock: [saveContentFx.done, saveBannersFx.done, saveLegalFx.done],
  target: contentInvalidated,
});

/* eslint-enable perfectionist/sort-objects */
