import { assertApiSuccess, parseApiErrorMessage } from './parseApiError';

export type Certificate = {
  created_at: string;
  file_url: string;
  id: string;
  is_active: boolean;
  sort_order: number;
  title: string;
  updated_at: string;
};

const parseCertificate = (value: unknown): Certificate | null => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;

  if (typeof record.id !== 'string' || record.id.length === 0 || typeof record.title !== 'string') {
    return null;
  }

  return {
    created_at: typeof record.created_at === 'string' ? record.created_at : '',
    file_url: typeof record.file_url === 'string' ? record.file_url : '',
    id: record.id,
    is_active: record.is_active === true,
    sort_order: typeof record.sort_order === 'number' ? record.sort_order : 0,
    title: record.title,
    updated_at: typeof record.updated_at === 'string' ? record.updated_at : '',
  };
};

export const parseCertificates = (data: unknown, fallback: string): Certificate[] => {
  const payload = assertApiSuccess(data, fallback).data;

  if (typeof payload !== 'object' || payload === null) {
    return [];
  }

  const items = (payload as { items?: unknown }).items;

  if (!Array.isArray(items)) {
    return [];
  }

  return items.flatMap((item) => {
    const certificate = parseCertificate(item);

    return certificate ? [certificate] : [];
  });
};

export const fetchCertificatesRequest = async (): Promise<Certificate[]> => {
  const response = await fetch('/api/v1/certificates');

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, 'Не удалось загрузить сертификаты'));
  }

  return parseCertificates(await response.json(), 'Некорректный ответ сертификатов');
};
