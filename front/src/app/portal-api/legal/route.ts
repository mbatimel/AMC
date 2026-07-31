import { apiOk } from '@/core/shared/server/portal/response';
import { readPortalState } from '@/core/shared/server/portal/store';

export const dynamic = 'force-dynamic';

export const GET = (): Response => apiOk({ items: readPortalState().legal_docs });
