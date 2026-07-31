import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import { appendAuditEntry, readPortalState, updatePortalState } from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type RouteContext = { params: Promise<{ id: string }> };

type UpdateLegalBody = {
  body: string;
  name?: string;
  summary?: string;
  version?: string;
};

export const GET = async (_request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const doc = readPortalState().legal_docs.find((item) => item.id === id);

  return doc ? apiOk(doc) : apiFail(404, 'Документ не найден');
};

export const PUT = async (request: Request, context: RouteContext): Promise<Response> => {
  const { id } = await context.params;
  const payload = await readJsonBody<UpdateLegalBody>(request);

  if (!payload) {
    return apiFail(400, 'Некорректное тело запроса');
  }

  let found = false;

  const state = updatePortalState((draft) => {
    const doc = draft.legal_docs.find((item) => item.id === id);

    if (!doc) {
      return;
    }

    found = true;
    doc.body = payload.body;
    doc.name = payload.name ?? doc.name;
    doc.updated_at = new Date().toISOString();

    if (payload.version && payload.version !== doc.current_version) {
      doc.current_version = payload.version;
      doc.versions.unshift({
        author: 'Админ портала',
        date: new Date().toLocaleDateString('ru-RU'),
        summary: payload.summary ?? 'Обновление документа',
        version: payload.version,
      });
    }
  });

  if (!found) {
    return apiFail(404, 'Документ не найден');
  }

  appendAuditEntry('Админ портала', `Обновлён юридический документ «${id}»`);

  return apiOk(state.legal_docs.find((item) => item.id === id));
};
