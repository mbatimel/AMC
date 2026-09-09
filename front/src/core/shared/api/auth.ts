import { parseApiErrorMessage } from './parseApiError';

export type AuthCredentials = {
  email: string;
  password: string;
};

export type AuthUserResponse = {
  userID: string;
};

export type RegisterIpPayload = {
  directorFullName: string;
  email: string;
  inn: string;
  password: string;
  phone: string;
  /** Файл с реквизитами организации (multipart field `requisitesFile`). */
  requisitesFile: File;
  shortName: string;
};

export class AuthApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'AuthApiError';
    this.status = status;
  }
}

const parseUserId = (data: unknown): string => {
  if (typeof data !== 'object' || data === null) {
    throw new AuthApiError(500, 'Некорректный ответ сервера');
  }

  const record = data as Record<string, unknown>;

  if (typeof record.userID === 'string' && record.userID.length > 0) {
    return record.userID;
  }

  if (typeof record.data === 'object' && record.data !== null) {
    const nested = record.data as Record<string, unknown>;

    if (typeof nested.userID === 'string' && nested.userID.length > 0) {
      return nested.userID;
    }
  }

  throw new AuthApiError(500, 'Некорректный ответ сервера');
};

const FIELD_LABELS: Record<string, string> = {
  email: 'E-mail',
  inn: 'ИНН',
  password: 'Пароль',
  requisitesFile: 'Файл с реквизитами',
};

type BlockedAccountDetails = {
  contactEmail?: string;
  contactName?: string;
  contactPhone?: string;
  reason?: string;
  siteDomain?: string;
};

const asOptionalString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined;
  }

  const trimmed = value.trim();

  return trimmed.length > 0 ? trimmed : undefined;
};

const parseBlockedAccountDetails = (value: unknown): BlockedAccountDetails | null => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;

  return {
    contactEmail: asOptionalString(record.contactEmail ?? record.contact_email),
    contactName: asOptionalString(record.contactName ?? record.contact_name),
    contactPhone: asOptionalString(record.contactPhone ?? record.contact_phone),
    reason: asOptionalString(record.reason),
    siteDomain: asOptionalString(record.siteDomain ?? record.site_domain),
  };
};

const formatBlockedAccountMessage = (details: BlockedAccountDetails): string => {
  const domain = details.siteDomain ?? (typeof window !== 'undefined' ? window.location.host : '');
  const parts = [
    domain
      ? `Ваша учётная запись была заблокирована администратором сайта ${domain}`
      : 'Ваша учётная запись была заблокирована администратором сайта',
  ];

  if (details.reason) {
    parts.push(`Причина блокировки: ${details.reason}`);
  }

  const contacts = [
    details.contactName ? `контактное лицо — ${details.contactName}` : null,
    details.contactPhone ? `телефон — ${details.contactPhone}` : null,
    details.contactEmail ? `e-mail — ${details.contactEmail}` : null,
  ].filter(Boolean);

  if (contacts.length > 0) {
    parts.push(`Контакты для связи: ${contacts.join(', ')}`);
  }

  return `${parts.join('. ')}.`;
};

const isBlockedAccountError = (message: string): boolean => {
  const normalized = message.trim().toLowerCase();

  return (
    normalized === 'user is blocked' ||
    normalized === 'user blocked' ||
    normalized.includes('account is blocked') ||
    normalized.includes('account blocked') ||
    normalized.includes('user is deactivated') ||
    normalized.includes('заблокирован')
  );
};

const localizeAuthError = (
  message: string,
  field?: string,
  blockedDetails?: BlockedAccountDetails | null,
): string => {
  const normalized = message.trim().toLowerCase();
  const fieldLabel = field ? (FIELD_LABELS[field] ?? field) : undefined;

  if (isBlockedAccountError(message)) {
    return formatBlockedAccountMessage(blockedDetails ?? {});
  }

  // Уже готовое сообщение с бэка — не перетираем.
  if (message.includes('учётная запись была заблокирована')) {
    return message.trim();
  }

  if (
    normalized === 'invalid email or password' ||
    normalized.includes('invalid email or password')
  ) {
    return 'Неверный email или пароль';
  }

  if (normalized === 'unauthorized' || normalized === 'unauthorised') {
    return 'Неверный email или пароль';
  }

  if (normalized === 'inn is empty' || normalized.startsWith('inn is empty')) {
    return 'Укажите ИНН';
  }

  if (normalized === 'inn is invalid' || normalized.startsWith('inn is invalid')) {
    return 'ИНН не найден или указан неверно';
  }

  if (normalized.includes('inn not valid') || normalized.includes('inn should be')) {
    return 'Введите корректный ИНН';
  }

  if (
    normalized.includes('requisites') ||
    normalized.includes('requisitesfile') ||
    normalized.includes('requisites file')
  ) {
    return 'Прикрепите корректный файл с реквизитами';
  }

  if (normalized === 'email already registered' || normalized.includes('email already')) {
    return 'Пользователь с таким email уже зарегистрирован';
  }

  if (
    normalized === 'auth.errors.tokeninvalid' ||
    normalized.includes('token invalid') ||
    normalized.includes('invalid or expired reset token') ||
    normalized.includes('invalid reset token')
  ) {
    return 'Ссылка для сброса пароля недействительна. Запросите новую.';
  }

  if (
    normalized === 'auth.errors.tokenexpired' ||
    normalized.includes('token expired') ||
    normalized.includes('expired reset token')
  ) {
    return 'Срок действия ссылки истёк. Запросите новую.';
  }

  if (normalized === 'validation failed' || normalized.startsWith('validation failed')) {
    return fieldLabel
      ? `Проверьте поле «${fieldLabel}»`
      : 'Проверьте правильность заполнения формы';
  }

  if (fieldLabel && normalized.includes(':')) {
    // «inn is empty: inn» и подобные сырые строки
    return localizeAuthError(normalized.split(':')[0] ?? normalized, undefined, blockedDetails);
  }

  return message;
};

const parseErrorMessage = async (response: Response): Promise<string> => {
  try {
    const data: unknown = await response.json();

    if (typeof data === 'object' && data !== null) {
      const record = data as Record<string, unknown>;
      const errorText = record.errorText ?? record.ErrorText;
      const additionalErrors = record.additionalErrors ?? record.AdditionalErrors;
      const blockedDetails = parseBlockedAccountDetails(additionalErrors);
      const cause = record.Cause ?? record.cause;
      let field: string | undefined;

      if (typeof cause === 'object' && cause !== null) {
        const causeRecord = cause as Record<string, unknown>;

        if (typeof causeRecord.field === 'string' && causeRecord.field.length > 0) {
          field = causeRecord.field;
        }
      }

      if (typeof errorText === 'string' && errorText.length > 0) {
        return localizeAuthError(errorText, field, blockedDetails);
      }

      if (typeof record.message === 'string' && record.message.length > 0) {
        return localizeAuthError(record.message, field, blockedDetails);
      }

      if (response.status === 403 && blockedDetails) {
        return formatBlockedAccountMessage(blockedDetails);
      }
    }
  } catch {
    // ignore JSON parse errors
  }

  if (response.status === 401) {
    return 'Неверный email или пароль';
  }

  if (response.status === 403) {
    return 'Доступ запрещён';
  }

  if (response.status === 409) {
    return 'Пользователь с таким email уже зарегистрирован';
  }

  if (response.status === 502 || response.status === 503 || response.status === 504) {
    return 'Сервис авторизации временно недоступен. Попробуйте позже';
  }

  return 'Не удалось выполнить запрос. Попробуйте позже';
};

const postAuth = async <TBody extends object>(
  path: string,
  body: TBody,
): Promise<AuthUserResponse> => {
  const response = await fetch(path, {
    body: JSON.stringify(body),
    headers: {
      'Content-Type': 'application/json',
    },
    method: 'POST',
  });

  if (!response.ok) {
    throw new AuthApiError(response.status, await parseErrorMessage(response));
  }

  const data: unknown = await response.json();

  return { userID: parseUserId(data) };
};

const postAuthMultipart = async (path: string, body: FormData): Promise<AuthUserResponse> => {
  // Content-Type намеренно не задаём — браузер сам проставит boundary.
  const response = await fetch(path, {
    body,
    method: 'POST',
  });

  if (!response.ok) {
    throw new AuthApiError(response.status, await parseErrorMessage(response));
  }

  const data: unknown = await response.json();

  return { userID: parseUserId(data) };
};

export const loginRequest = async ({
  email,
  password,
}: AuthCredentials): Promise<AuthUserResponse> => {
  // Swagger: POST /api/v1/auth/login — JSON body { email, password }
  return postAuth('/api/v1/auth/login', { email, password });
};

/**
 * Регистрация ИП/организации.
 * Контракт: POST /api/v1/auth/register/ip — multipart/form-data
 * обязательные поля: email, password, shortName, inn, directorFullName, phone, requisitesFile.
 */
export const registerIpRequest = async (payload: RegisterIpPayload): Promise<AuthUserResponse> => {
  const body = new FormData();

  body.set('email', payload.email);
  body.set('password', payload.password);
  body.set('shortName', payload.shortName);
  body.set('inn', payload.inn);
  body.set('directorFullName', payload.directorFullName);
  body.set('phone', payload.phone);
  body.set('requisitesFile', payload.requisitesFile, payload.requisitesFile.name);

  return postAuthMultipart('/api/v1/auth/register/ip', body);
};

export const changePasswordRequest = async (params: {
  newPassword: string;
  oldPassword: string;
  userId: string;
}): Promise<void> => {
  // Swagger: POST /api/v1/auth/change-password
  // header X-User-Id; JSON body { oldPassword, newPassword }
  const response = await fetch('/api/v1/auth/change-password', {
    body: JSON.stringify({
      newPassword: params.newPassword,
      oldPassword: params.oldPassword,
    }),
    headers: {
      'Content-Type': 'application/json',
      'X-User-Id': params.userId,
    },
    method: 'POST',
  });

  if (!response.ok) {
    if (response.status === 401) {
      throw new AuthApiError(response.status, 'Неверный текущий пароль');
    }

    throw new AuthApiError(
      response.status,
      await parseApiErrorMessage(response, 'Не удалось сменить пароль'),
    );
  }
};

export type ForgotPasswordResult = {
  emailSent: boolean;
};

/**
 * Запрос ссылки для сброса пароля (публичный, без сессии).
 * Контракт: POST /api/v1/auth/password/reset/request — JSON { email }.
 * Ответ всегда успешный при валидном email; наличие аккаунта не раскрывается.
 */
export const requestPasswordResetRequest = async (email: string): Promise<ForgotPasswordResult> => {
  const response = await fetch('/api/v1/auth/password/reset/request', {
    body: JSON.stringify({ email }),
    headers: {
      'Content-Type': 'application/json',
    },
    method: 'POST',
  });

  if (!response.ok) {
    throw new AuthApiError(response.status, await parseErrorMessage(response));
  }

  try {
    const data: unknown = await response.json();

    if (typeof data === 'object' && data !== null) {
      const record = data as Record<string, unknown>;
      const payload =
        typeof record.data === 'object' && record.data !== null
          ? (record.data as Record<string, unknown>)
          : record;

      if (typeof payload.emailSent === 'boolean') {
        return { emailSent: payload.emailSent };
      }
    }
  } catch {
    // ignore JSON parse errors — считаем успехом
  }

  return { emailSent: true };
};

/**
 * Установка нового пароля по токену из письма (публичный, без сессии).
 * Контракт: POST /api/v1/auth/password/reset/confirm — JSON { token, newPassword }.
 */
export const confirmPasswordResetRequest = async (params: {
  newPassword: string;
  token: string;
}): Promise<void> => {
  const response = await fetch('/api/v1/auth/password/reset/confirm', {
    body: JSON.stringify({
      newPassword: params.newPassword,
      token: params.token,
    }),
    headers: {
      'Content-Type': 'application/json',
    },
    method: 'POST',
  });

  if (!response.ok) {
    throw new AuthApiError(response.status, await parseErrorMessage(response));
  }
};
