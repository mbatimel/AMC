import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import { appendAuditEntry, readPortalState } from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

type CreateAuditBody = {
  action?: string;
  actorLabel?: string;
};

export const GET = (request: Request): Response => {
  const params = new URL(request.url).searchParams;
  const limit = Math.min(Number(params.get('limit')) || 50, 200);
  const offset = Number(params.get('offset')) || 0;
  const all = readPortalState().audit_log;

  return apiOk({
    items: all.slice(offset, offset + limit),
    pagination: { limit, offset, total: all.length },
  });
};

export const POST = async (request: Request): Promise<Response> => {
  const body = await readJsonBody<CreateAuditBody>(request);

  if (!body?.action?.trim()) {
    return apiFail(400, 'Не указано действие');
  }

  appendAuditEntry(body.actorLabel?.trim() || 'Админ портала', body.action.trim());

  return apiOk({ ok: true });
};
