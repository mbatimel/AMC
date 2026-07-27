export type AuthCredentials = {
  email: string;
  password: string;
};

export type AuthUserResponse = {
  userID: string;
};

export type SignUpPayload = AuthCredentials & {
  name: string;
  surename: string;
};

export class AuthApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'AuthApiError';
    this.status = status;
  }
}

const buildUrl = (path: string, params: Record<string, string>): string => {
  const searchParams = new URLSearchParams(params);

  return `${path}?${searchParams.toString()}`;
};

const parseErrorMessage = async (response: Response): Promise<string> => {
  try {
    const data: unknown = await response.json();

    if (typeof data === 'object' && data !== null) {
      const record = data as Record<string, unknown>;

      if (typeof record.errorText === 'string' && record.errorText.length > 0) {
        return record.errorText;
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

  if (response.status === 502 || response.status === 503 || response.status === 504) {
    return 'Сервис авторизации временно недоступен. Попробуйте позже';
  }

  return 'Не удалось выполнить запрос. Попробуйте позже';
};

const postAuth = async (
  path: string,
  params: Record<string, string>,
): Promise<AuthUserResponse> => {
  const response = await fetch(buildUrl(path, params), {
    method: 'POST',
  });

  if (!response.ok) {
    throw new AuthApiError(response.status, await parseErrorMessage(response));
  }

  const data: unknown = await response.json();

  if (typeof data !== 'object' || data === null || !('userID' in data)) {
    throw new AuthApiError(500, 'Некорректный ответ сервера');
  }

  const userID = (data as AuthUserResponse).userID;

  if (typeof userID !== 'string' || userID.length === 0) {
    throw new AuthApiError(500, 'Некорректный ответ сервера');
  }

  return { userID };
};

export const loginRequest = async ({
  email,
  password,
}: AuthCredentials): Promise<AuthUserResponse> => {
  return postAuth('/api/v1/auth/login', { email, password });
};

export const signupRequest = async ({
  email,
  name,
  password,
  surename,
}: SignUpPayload): Promise<AuthUserResponse> => {
  return postAuth('/api/v1/auth/signup', { email, name, password, surename });
};
