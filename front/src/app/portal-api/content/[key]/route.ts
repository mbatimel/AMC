import type { ContentPageKey } from '@/core/shared/server/portal/types';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  appendAuditEntry,
  readPortalState,
  updatePortalState,
} from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

const PAGE_KEYS: ContentPageKey[] = [
  'about',
  'certificates',
  'contacts',
  'home',
  'promo',
  'terms',
];

const isPageKey = (value: string): value is ContentPageKey =>
  PAGE_KEYS.includes(value as ContentPageKey);

type RouteContext = { params: Promise<{ key: string }> };

export const GET = async (_request: Request, context: RouteContext): Promise<Response> => {
  const { key } = await context.params;

  if (!isPageKey(key)) {
    return apiFail(404, 'Раздел контента не найден');
  }

  return apiOk(readPortalState().content[key]);
};

export const PUT = async (request: Request, context: RouteContext): Promise<Response> => {
  const { key } = await context.params;

  if (!isPageKey(key)) {
    return apiFail(404, 'Раздел контента не найден');
  }

  const body = await readJsonBody<Record<string, unknown>>(request);

  if (!body) {
    return apiFail(400, 'Некорректное тело запроса');
  }

  updatePortalState((draft) => {
    const pages = draft.content as unknown as Record<string, Record<string, unknown>>;

    pages[key] = { ...pages[key], ...body };
  });

  appendAuditEntry('Админ портала', `Обновлён раздел контента «${key}»`);

  return apiOk(readPortalState().content[key]);
};
