import type {
  BannerItem,
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

export const fetchBannersRequest = (): Promise<BannersSettings> =>
  portalRequest({ fallback: 'Не удалось загрузить баннеры', path: '/banners' });

export const updateBannersRequest = (payload: BannersSettings): Promise<BannersSettings> =>
  portalRequest({
    body: payload,
    fallback: 'Не удалось сохранить баннеры',
    method: 'PUT',
    path: '/banners',
  });

const bannerAdminRequest = async <T>(
  userId: string,
  path: string,
  options: RequestInit = {},
  fallback = 'Не удалось выполнить операцию с баннером',
): Promise<T> => {
  const headers = new Headers(options.headers);

  headers.set('X-User-Id', userId);
  const response = await fetch(`/api/v1/admin/banners${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, fallback));
  }

  return assertApiSuccess(await response.json(), fallback).data as T;
};

export const fetchAdminBannersRequest = (userId: string): Promise<BannersSettings> =>
  bannerAdminRequest(userId, '', {}, 'Не удалось загрузить баннеры');

const bannerForm = (item: BannerItem, file?: File): FormData => {
  const body = new FormData();

  body.set('title', item.title);
  body.set('subtitle', item.subtitle);
  body.set('link', item.link);
  body.set('sort_order', String(item.sort_order));
  body.set('is_active', String(item.is_active));
  body.set('dateFrom', item.dateFrom);
  body.set('dateTo', item.dateTo);
  if (file) {
    body.set('file', file);
  }

  return body;
};

export const createBannerRequest = (
  userId: string,
  item: BannerItem,
  file: File,
): Promise<BannerItem> =>
  bannerAdminRequest(userId, '', { body: bannerForm(item, file), method: 'POST' });

export const updateBannerRequest = (
  userId: string,
  item: BannerItem,
  file?: File,
): Promise<BannerItem> =>
  bannerAdminRequest(userId, `/${item.id}`, {
    body: bannerForm(item, file),
    method: 'PATCH',
  });

export const deleteBannerRequest = (userId: string, bannerId: string): Promise<void> =>
  bannerAdminRequest(userId, `/${bannerId}`, { method: 'DELETE' });

export const updateBannerDelayRequest = (
  userId: string,
  delaySec: number,
): Promise<BannersSettings> =>
  bannerAdminRequest(userId, '/settings', {
    body: JSON.stringify({ delay_sec: delaySec }),
    headers: { 'Content-Type': 'application/json' },
    method: 'PUT',
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
