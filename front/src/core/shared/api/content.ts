import type {
  BannersSettings,
  ContentPageKey,
  ContentPages,
} from '@/core/shared/server/portal/types';

import { assertApiSuccess, parseApiErrorMessage } from './parseApiError';
import { portalRequest } from './portalClient';

export type { LegalDoc, LegalDocVersion } from './legalDocs';
export { fetchLegalDocsRequest } from './legalDocs';
export type {
  AboutPageContent,
  BannerItem,
  BannersSettings,
  ContactOffice,
  ContactRequisiteItem,
  ContactsPageContent,
  ContentPageKey,
  ContentPages,
  HomePageContent,
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
