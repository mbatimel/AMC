import { assertApiSuccess, parseApiErrorMessage } from './parseApiError';

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

export const parseLegalDocs = (data: unknown, fallback: string): LegalDoc[] => {
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

export const fetchLegalDocsRequest = async (): Promise<LegalDoc[]> => {
  const response = await fetch('/api/v1/legal-docs');

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, 'Не удалось загрузить документы'));
  }

  return parseLegalDocs(await response.json(), 'Некорректный ответ документов');
};
