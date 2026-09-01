import { createEffect, createEvent, createStore, sample } from 'effector';

import type { Certificate } from '@/core/shared/api/certificates';
import type { BannersSettings, ContentPages, LegalDoc } from '@/core/shared/api/content';

import { fetchCertificatesRequest } from '@/core/shared/api/certificates';
import {
  fetchBannersRequest,
  fetchContentPagesRequest,
  fetchLegalDocsRequest,
} from '@/core/shared/api/content';
import { toDisplayErrorMessage } from '@/core/shared/api/parseApiError';

export const contentRequested = createEvent();
export const contentInvalidated = createEvent();

export const fetchContentFx = createEffect(() => fetchContentPagesRequest());
export const fetchBannersFx = createEffect(() => fetchBannersRequest());
export const fetchLegalDocsFx = createEffect(() => fetchLegalDocsRequest());
export const fetchCertificatesFx = createEffect(() => fetchCertificatesRequest());

export const $content = createStore<ContentPages | null>(null).on(
  fetchContentFx.doneData,
  (_, content) => content,
);

export const $banners = createStore<BannersSettings | null>(null).on(
  fetchBannersFx.doneData,
  (_, banners) => banners,
);

export const $legalDocs = createStore<LegalDoc[]>([]).on(
  fetchLegalDocsFx.doneData,
  (_, docs) => docs,
);

export const $certificates = createStore<Certificate[]>([]).on(
  fetchCertificatesFx.doneData,
  (_, items) => items,
);

export const $isContentPending = fetchContentFx.pending;

export const $contentError = createStore<null | string>(null)
  .on(fetchContentFx, () => null)
  .on(fetchContentFx.failData, (_, error) =>
    toDisplayErrorMessage(error, 'Не удалось загрузить контент портала'),
  );

const $isContentLoaded = createStore(false)
  .on(fetchContentFx.done, () => true)
  .reset(contentInvalidated);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> target */

sample({
  clock: contentRequested,
  source: { loaded: $isContentLoaded, pending: fetchContentFx.pending },
  filter: ({ loaded, pending }) => !loaded && !pending,
  target: [fetchContentFx, fetchBannersFx, fetchLegalDocsFx, fetchCertificatesFx],
});

sample({
  clock: contentInvalidated,
  target: [fetchContentFx, fetchBannersFx, fetchLegalDocsFx, fetchCertificatesFx],
});

/* eslint-enable perfectionist/sort-objects */
