import { assertApiSuccess, parseApiErrorMessage } from './parseApiError';

export type CreateLegalDocPayload = {
  fileContentBase64: string;
  fileName: string;
  id: string;
  name: string;
  summary: string;
  version: string;
};

export type LegalDoc = {
  current_version: string;
  file_url: string;
  id: string;
  name: string;
  updated_at: string;
};

export type LegalDocVersion = {
  author: string;
  created_at: string;
  file_url: string;
  id: string;
  summary: string;
  version: string;
};

export type ReplaceLegalDocFilePayload = {
  docId: string;
  fileContentBase64: string;
  fileName: string;
  summary: string;
  version: string;
};

const parseLegalDoc = (value: unknown): LegalDoc | null => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;

  if (typeof record.id !== 'string' || record.id.length === 0 || typeof record.name !== 'string') {
    return null;
  }

  return {
    current_version: typeof record.current_version === 'string' ? record.current_version : '',
    file_url: typeof record.file_url === 'string' ? record.file_url : '',
    id: record.id,
    name: record.name,
    updated_at: typeof record.updated_at === 'string' ? record.updated_at : '',
  };
};

const parseLegalDocVersion = (value: unknown): LegalDocVersion | null => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;

  if (typeof record.version !== 'string') {
    return null;
  }

  return {
    author: typeof record.author === 'string' ? record.author : '',
    created_at: typeof record.created_at === 'string' ? record.created_at : '',
    file_url: typeof record.file_url === 'string' ? record.file_url : '',
    id: typeof record.id === 'string' ? record.id : record.version,
    summary: typeof record.summary === 'string' ? record.summary : '',
    version: record.version,
  };
};

const parseLegalDocs = (data: unknown, fallback: string): LegalDoc[] => {
  const payload = assertApiSuccess(data, fallback).data;

  if (typeof payload !== 'object' || payload === null) {
    return [];
  }

  const items = (payload as { items?: unknown }).items;

  if (!Array.isArray(items)) {
    return [];
  }

  return items.flatMap((item) => {
    const doc = parseLegalDoc(item);

    return doc ? [doc] : [];
  });
};

const legalAdminRequest = async (
  userId: string,
  path: string,
  options: RequestInit = {},
  fallback = 'Не удалось выполнить операцию с документом',
): Promise<unknown> => {
  const headers = new Headers(options.headers);

  headers.set('X-User-Id', userId);
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(`/api/v1/admin/legal-docs${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, fallback));
  }

  return assertApiSuccess(await response.json(), fallback).data;
};

export const fetchLegalDocsRequest = async (): Promise<LegalDoc[]> => {
  const response = await fetch('/api/v1/legal-docs');

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, 'Не удалось загрузить документы'));
  }

  return parseLegalDocs(await response.json(), 'Некорректный ответ документов');
};

export const fetchAdminLegalDocsRequest = async (userId: string): Promise<LegalDoc[]> => {
  const payload = await legalAdminRequest(userId, '', {}, 'Не удалось загрузить документы');

  if (typeof payload !== 'object' || payload === null) {
    return [];
  }

  const items = (payload as { items?: unknown }).items;

  if (!Array.isArray(items)) {
    return [];
  }

  return items.flatMap((item) => {
    const doc = parseLegalDoc(item);

    return doc ? [doc] : [];
  });
};

export const createLegalDocRequest = async (
  userId: string,
  payload: CreateLegalDocPayload,
): Promise<LegalDoc> => {
  const data = await legalAdminRequest(
    userId,
    '',
    {
      body: JSON.stringify(payload),
      method: 'POST',
    },
    'Не удалось добавить документ',
  );
  const doc = parseLegalDoc(data);

  if (!doc) {
    throw new Error('Некорректный ответ создания документа');
  }

  return doc;
};

export const replaceLegalDocFileRequest = async (
  userId: string,
  payload: ReplaceLegalDocFilePayload,
): Promise<LegalDoc> => {
  const data = await legalAdminRequest(
    userId,
    `/${encodeURIComponent(payload.docId)}`,
    {
      body: JSON.stringify({
        fileContentBase64: payload.fileContentBase64,
        fileName: payload.fileName,
        summary: payload.summary,
        version: payload.version,
      }),
      method: 'PATCH',
    },
    'Не удалось заменить файл документа',
  );
  const doc = parseLegalDoc(data);

  if (!doc) {
    throw new Error('Некорректный ответ замены файла');
  }

  return doc;
};

export const deleteLegalDocRequest = async (userId: string, docId: string): Promise<void> => {
  await legalAdminRequest(
    userId,
    `/${encodeURIComponent(docId)}`,
    { method: 'DELETE' },
    'Не удалось удалить документ',
  );
};

export const fetchLegalDocVersionsRequest = async (
  userId: string,
  docId: string,
): Promise<LegalDocVersion[]> => {
  const payload = await legalAdminRequest(
    userId,
    `/${encodeURIComponent(docId)}/versions`,
    {},
    'Не удалось загрузить историю версий',
  );

  if (typeof payload !== 'object' || payload === null) {
    return [];
  }

  const items = (payload as { items?: unknown }).items;

  if (!Array.isArray(items)) {
    return [];
  }

  return items.flatMap((item) => {
    const version = parseLegalDocVersion(item);

    return version ? [version] : [];
  });
};
