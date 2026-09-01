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

export type CreateCertificatePayload = {
  fileContentBase64: string;
  fileName: string;
  isActive: boolean;
  sortOrder: number;
  title: string;
};

export type UpdateCertificatePayload = {
  certId: string;
  fileContentBase64?: string;
  fileName?: string;
  isActive: boolean;
  sortOrder: number;
  title: string;
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

const certificatesAdminRequest = async (
  userId: string,
  path: string,
  options: RequestInit = {},
  fallback = 'Не удалось выполнить операцию с сертификатом',
): Promise<unknown> => {
  const headers = new Headers(options.headers);

  headers.set('X-User-Id', userId);
  if (options.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(`/api/v1/admin/certificates${path}`, {
    ...options,
    headers,
  });

  if (!response.ok) {
    throw new Error(await parseApiErrorMessage(response, fallback));
  }

  return assertApiSuccess(await response.json(), fallback).data;
};

export const fetchAdminCertificatesRequest = async (userId: string): Promise<Certificate[]> => {
  const payload = await certificatesAdminRequest(
    userId,
    '',
    {},
    'Не удалось загрузить сертификаты',
  );

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

export const createCertificateRequest = async (
  userId: string,
  payload: CreateCertificatePayload,
): Promise<Certificate> => {
  const data = await certificatesAdminRequest(
    userId,
    '',
    {
      body: JSON.stringify(payload),
      method: 'POST',
    },
    'Не удалось добавить сертификат',
  );
  const certificate = parseCertificate(data);

  if (!certificate) {
    throw new Error('Некорректный ответ создания сертификата');
  }

  return certificate;
};

export const updateCertificateRequest = async (
  userId: string,
  payload: UpdateCertificatePayload,
): Promise<Certificate> => {
  const data = await certificatesAdminRequest(
    userId,
    `/${encodeURIComponent(payload.certId)}`,
    {
      body: JSON.stringify({
        fileContentBase64: payload.fileContentBase64 ?? '',
        fileName: payload.fileName ?? '',
        isActive: payload.isActive,
        sortOrder: payload.sortOrder,
        title: payload.title,
      }),
      method: 'PATCH',
    },
    'Не удалось сохранить сертификат',
  );
  const certificate = parseCertificate(data);

  if (!certificate) {
    throw new Error('Некорректный ответ сохранения сертификата');
  }

  return certificate;
};

export const deleteCertificateRequest = async (userId: string, certId: string): Promise<void> => {
  await certificatesAdminRequest(
    userId,
    `/${encodeURIComponent(certId)}`,
    { method: 'DELETE' },
    'Не удалось удалить сертификат',
  );
};
