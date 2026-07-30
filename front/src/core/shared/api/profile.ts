import { assertApiSuccess, parseApiErrorMessage } from './parseApiError';

export type CategoryDiscount = {
  category_id: string;
  category_name: string;
  discount_percent: number;
  valid_from?: string;
  valid_to?: string;
};

export type ClientConditions = {
  category_discounts: CategoryDiscount[];
  client_id: string;
  contact_channel: string;
  credit_limit: number;
  credit_used: number;
  payment_terms: string;
  price_group: string;
  sales_contact: string;
};

export type Profile = {
  active_client: null | ProfileClient;
  active_client_id: string;
  email: string;
  first_name: string;
  is_active: boolean;
  last_name: string;
  middle_name: string;
  phone: string;
  status: string;
  user_id: string;
};

export type ProfileClient = {
  address: string;
  company_name: string;
  company_type: string;
  contact_name: string;
  email: string;
  id: string;
  inn: string;
  ogrn: string;
  phone: string;
};

export type ProfileClientListItem = {
  client: ProfileClient;
  is_active: boolean;
};

export type UpdateProfilePayload = {
  email: string;
  firstName: string;
  lastName: string;
  middleName: string;
  phone: string;
};

export class ProfileApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'ProfileApiError';
    this.status = status;
  }
}

const asNumber = (value: unknown, fallback = 0): number =>
  typeof value === 'number' && Number.isFinite(value) ? value : fallback;

const asString = (value: unknown, fallback = ''): string =>
  typeof value === 'string' ? value : fallback;

const asBoolean = (value: unknown, fallback = false): boolean =>
  typeof value === 'boolean' ? value : fallback;

const withUserHeaders = (userId: string, json = false): HeadersInit => ({
  ...(json ? { 'Content-Type': 'application/json' } : {}),
  'X-User-Id': userId,
});

const throwIfNotOk = async (response: Response, fallback: string): Promise<void> => {
  if (!response.ok) {
    throw new ProfileApiError(response.status, await parseApiErrorMessage(response, fallback));
  }
};

const parseClient = (value: unknown): null | ProfileClient => {
  if (typeof value !== 'object' || value === null) {
    return null;
  }

  const record = value as Record<string, unknown>;
  const id = asString(record.id);

  if (!id) {
    return null;
  }

  return {
    address: asString(record.address),
    company_name: asString(record.company_name),
    company_type: asString(record.company_type),
    contact_name: asString(record.contact_name),
    email: asString(record.email),
    id,
    inn: asString(record.inn),
    ogrn: asString(record.ogrn),
    phone: asString(record.phone),
  };
};

const parseProfile = (data: unknown): Profile => {
  const record = assertApiSuccess(data, 'Некорректный ответ профиля');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new ProfileApiError(500, 'Некорректный ответ профиля');
  }

  const payloadRecord = payload as Record<string, unknown>;
  const profilePayload = payloadRecord.profile ?? payload;

  if (typeof profilePayload !== 'object' || profilePayload === null) {
    throw new ProfileApiError(500, 'Некорректный ответ профиля');
  }

  const profile = profilePayload as Record<string, unknown>;

  return {
    active_client: parseClient(profile.active_client),
    active_client_id: asString(profile.active_client_id),
    email: asString(profile.email),
    first_name: asString(profile.first_name),
    is_active: asBoolean(profile.is_active, true),
    last_name: asString(profile.last_name),
    middle_name: asString(profile.middle_name),
    phone: asString(profile.phone),
    status: asString(profile.status),
    user_id: asString(profile.user_id),
  };
};

export const getProfileRequest = async (userId: string): Promise<Profile> => {
  const response = await fetch('/api/v1/profile', {
    headers: withUserHeaders(userId),
  });

  await throwIfNotOk(response, 'Не удалось загрузить профиль');

  return parseProfile(await response.json());
};

export const updateProfileRequest = async (
  userId: string,
  payload: UpdateProfilePayload,
): Promise<Profile> => {
  const response = await fetch('/api/v1/profile', {
    body: JSON.stringify(payload),
    headers: withUserHeaders(userId, true),
    method: 'PATCH',
  });

  await throwIfNotOk(response, 'Не удалось обновить профиль');

  return parseProfile(await response.json());
};

export const listUserClientsRequest = async (userId: string): Promise<ProfileClientListItem[]> => {
  const response = await fetch('/api/v1/profile/clients', {
    headers: withUserHeaders(userId),
  });

  await throwIfNotOk(response, 'Не удалось загрузить список кабинетов');

  const record = assertApiSuccess(await response.json(), 'Некорректный ответ кабинетов');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    return [];
  }

  const items = (payload as Record<string, unknown>).items;

  if (!Array.isArray(items)) {
    return [];
  }

  return items.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const row = item as Record<string, unknown>;
    const client = parseClient(row.client);

    if (!client) {
      return [];
    }

    return [{ client, is_active: asBoolean(row.is_active) }];
  });
};

export const getClientDetailsRequest = async (
  userId: string,
  clientID: string,
): Promise<ProfileClient> => {
  const response = await fetch(`/api/v1/profile/clients/${clientID}`, {
    headers: withUserHeaders(userId),
  });

  await throwIfNotOk(response, 'Не удалось загрузить реквизиты');

  const record = assertApiSuccess(await response.json(), 'Некорректный ответ реквизитов');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new ProfileApiError(500, 'Некорректный ответ реквизитов');
  }

  const details = (payload as Record<string, unknown>).details;
  const clientPayload =
    typeof details === 'object' && details !== null
      ? (details as Record<string, unknown>).client
      : null;
  const client = parseClient(clientPayload);

  if (!client) {
    throw new ProfileApiError(500, 'Некорректный ответ реквизитов');
  }

  return client;
};

export const getClientConditionsRequest = async (
  userId: string,
  clientID: string,
): Promise<ClientConditions> => {
  const response = await fetch(`/api/v1/profile/clients/${clientID}/conditions`, {
    headers: withUserHeaders(userId),
  });

  await throwIfNotOk(response, 'Не удалось загрузить условия');

  const record = assertApiSuccess(await response.json(), 'Некорректный ответ условий');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new ProfileApiError(500, 'Некорректный ответ условий');
  }

  const conditionsPayload = (payload as Record<string, unknown>).conditions ?? payload;

  if (typeof conditionsPayload !== 'object' || conditionsPayload === null) {
    throw new ProfileApiError(500, 'Некорректный ответ условий');
  }

  const conditions = conditionsPayload as Record<string, unknown>;
  const rawDiscounts = Array.isArray(conditions.category_discounts)
    ? conditions.category_discounts
    : [];

  return {
    category_discounts: rawDiscounts.flatMap((item) => {
      if (typeof item !== 'object' || item === null) {
        return [];
      }

      const discount = item as Record<string, unknown>;

      return [
        {
          category_id: asString(discount.category_id),
          category_name: asString(discount.category_name),
          discount_percent: asNumber(discount.discount_percent),
          valid_from: asString(discount.valid_from) || undefined,
          valid_to: asString(discount.valid_to) || undefined,
        },
      ];
    }),
    client_id: asString(conditions.client_id, clientID),
    contact_channel: asString(conditions.contact_channel),
    credit_limit: asNumber(conditions.credit_limit),
    credit_used: asNumber(conditions.credit_used),
    payment_terms: asString(conditions.payment_terms),
    price_group: asString(conditions.price_group),
    sales_contact: asString(conditions.sales_contact),
  };
};

export const activateClientRequest = async (
  userId: string,
  clientID: string,
): Promise<ProfileClientListItem> => {
  const response = await fetch(`/api/v1/profile/clients/${clientID}/activate`, {
    headers: withUserHeaders(userId),
    method: 'POST',
  });

  await throwIfNotOk(response, 'Не удалось переключить кабинет');

  const record = assertApiSuccess(await response.json(), 'Некорректный ответ активации');
  const payload = record.data;

  if (typeof payload !== 'object' || payload === null) {
    throw new ProfileApiError(500, 'Некорректный ответ активации');
  }

  const active = (payload as Record<string, unknown>).active_client;

  if (typeof active !== 'object' || active === null) {
    throw new ProfileApiError(500, 'Некорректный ответ активации');
  }

  const row = active as Record<string, unknown>;
  const client = parseClient(row.client);

  if (!client) {
    throw new ProfileApiError(500, 'Некорректный ответ активации');
  }

  return { client, is_active: asBoolean(row.is_active, true) };
};
