export type AuthCredentials = {
  email: string;
  password: string;
};

export type AuthUserResponse = {
  userID: string;
};

export type RegisterIndividualPayload = AuthCredentials & {
  city: string;
  deliveryAddress: string;
  fio: string;
  inn?: string;
  phone: string;
};

export type RegisterIpPayload = AuthCredentials & {
  actualAddress?: string;
  additionalPhone?: string;
  bankAccount?: string;
  bankBik?: string;
  bankName?: string;
  correspondentAccount?: string;
  directorFullName?: string;
  directorPosition?: string;
  fullName?: string;
  inn?: string;
  kpp?: string;
  legalAddress?: string;
  ogrn?: string;
  okved?: string;
  phone?: string;
  shortName?: string;
  taxSystem?: string;
  website?: string;
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

const parseErrorMessage = async (response: Response): Promise<string> => {
  try {
    const data: unknown = await response.json();

    if (typeof data === 'object' && data !== null) {
      const record = data as Record<string, unknown>;
      const errorText = record.errorText ?? record.ErrorText;

      if (typeof errorText === 'string' && errorText.length > 0) {
        const cause = record.Cause ?? record.cause;

        if (typeof cause === 'object' && cause !== null && 'field' in cause) {
          const field = (cause as { field?: unknown }).field;

          if (typeof field === 'string' && field.length > 0) {
            return `${errorText}: ${field}`;
          }
        }

        return errorText;
      }

      if (typeof record.message === 'string' && record.message.length > 0) {
        return record.message;
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

const omitEmptyFields = <T extends Record<string, string | undefined>>(payload: T): Partial<T> => {
  return Object.fromEntries(
    Object.entries(payload).filter(([, value]) => value !== undefined && value.trim() !== ''),
  ) as Partial<T>;
};

export const loginRequest = async ({
  email,
  password,
}: AuthCredentials): Promise<AuthUserResponse> => {
  return postAuth('/api/v1/auth/login', { email, password });
};

export const registerIpRequest = async ({
  email,
  password,
  ...optionalFields
}: RegisterIpPayload): Promise<AuthUserResponse> => {
  return postAuth('/api/v1/auth/register/ip', {
    email,
    password,
    ...omitEmptyFields(optionalFields),
  });
};

export const registerIndividualRequest = async (
  payload: RegisterIndividualPayload,
): Promise<AuthUserResponse> => {
  const { city, deliveryAddress, email, fio, inn, password, phone } = payload;

  return postAuth('/api/v1/auth/register/individual', {
    city,
    deliveryAddress,
    email,
    fio,
    password,
    phone,
    ...omitEmptyFields({ inn }),
  });
};
