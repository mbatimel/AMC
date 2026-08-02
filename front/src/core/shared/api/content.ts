import type {
  BannersSettings,
  ContentPageKey,
  ContentPages,
  LegalDoc,
} from '@/core/shared/server/portal/types';

import { assertApiSuccess, parseApiErrorMessage } from './parseApiError';
import { portalRequest } from './portalClient';

export type {
  BannerItem,
  BannersSettings,
  ContactsPageContent,
  ContentPageKey,
  ContentPages,
  HomePageContent,
  LegalDoc,
  ListPageContent,
  TextPageContent,
} from '@/core/shared/server/portal/types';

export const CONTENT_PAGE_TITLES: Record<ContentPageKey, string> = {
  about: 'О компании',
  certificates: 'Сертификаты',
  contacts: 'Контакты',
  home: 'Главная страница',
  promo: 'Акции',
  terms: 'Условия работы',
};

export const fetchContentPagesRequest = (): Promise<ContentPages> =>
  portalRequest({ fallback: 'Не удалось загрузить контент портала', path: '/content' });

export const fetchContentPageRequest = <K extends ContentPageKey>(
  key: K,
): Promise<ContentPages[K]> =>
  portalRequest({ fallback: 'Не удалось загрузить страницу', path: `/content/${key}` });

export const updateContentPageRequest = <K extends ContentPageKey>(
  key: K,
  payload: ContentPages[K],
): Promise<ContentPages[K]> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось сохранить страницу',
    method: 'PUT',
    path: `/content/${key}`,
  });

export const fetchBannersRequest = async (): Promise<BannersSettings> => {
  const response = await fetch('/api/v1/banners');

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, 'Не удалось загрузить баннеры'));
  }

  return assertApiSuccess(await response.json(), 'Не удалось загрузить баннеры')
    .data as BannersSettings;
};

export const updateBannersRequest = (payload: BannersSettings): Promise<BannersSettings> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось сохранить баннеры',
    method: 'PUT',
    path: '/banners',
  });

export const fetchLegalDocsRequest = async (): Promise<LegalDoc[]> => {
  const result = await portalRequest<{ items: LegalDoc[] }>({
    fallback: 'Не удалось загрузить юридические документы',
    path: '/legal',
  });

  return result.items;
};

export const fetchLegalDocRequest = (docId: string): Promise<LegalDoc> =>
  portalRequest({ fallback: 'Не удалось загрузить документ', path: `/legal/${docId}` });

export const updateLegalDocRequest = (
  docId: string,
  payload: { body: string; name?: string; summary?: string; version?: string },
): Promise<LegalDoc> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось сохранить документ',
    method: 'PUT',
    path: `/legal/${docId}`,
  });
