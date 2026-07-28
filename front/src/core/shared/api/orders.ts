export type CityItem = {
  city: string;
  id: string;
};

export class OrdersApiError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'OrdersApiError';
    this.status = status;
  }
}

const parseCities = (data: unknown): CityItem[] => {
  if (typeof data !== 'object' || data === null) {
    throw new OrdersApiError(500, 'Некорректный ответ сервера');
  }

  const record = data as Record<string, unknown>;

  if (record.error === true) {
    const errorText = record.errorText ?? record.ErrorText;

    throw new OrdersApiError(
      500,
      typeof errorText === 'string' && errorText.length > 0
        ? errorText
        : 'Не удалось загрузить список городов',
    );
  }

  const cities = record.data;

  if (!Array.isArray(cities)) {
    throw new OrdersApiError(500, 'Некорректный ответ сервера');
  }

  return cities.flatMap((item) => {
    if (typeof item !== 'object' || item === null) {
      return [];
    }

    const cityRecord = item as Record<string, unknown>;
    const id = cityRecord.id;
    const city = cityRecord.city;

    if (
      typeof id !== 'string' ||
      id.length === 0 ||
      typeof city !== 'string' ||
      city.length === 0
    ) {
      return [];
    }

    return [{ city, id }];
  });
};

const parseErrorMessage = async (response: Response): Promise<string> => {
  try {
    const data: unknown = await response.json();

    if (typeof data === 'object' && data !== null) {
      const record = data as Record<string, unknown>;
      const errorText = record.errorText ?? record.ErrorText;

      if (typeof errorText === 'string' && errorText.length > 0) {
        return errorText;
      }
    }
  } catch {
    // ignore JSON parse errors
  }

  if (response.status === 502 || response.status === 503 || response.status === 504) {
    return 'Сервис временно недоступен. Попробуйте позже';
  }

  return 'Не удалось загрузить список городов';
};

export const getCitiesRequest = async (): Promise<CityItem[]> => {
  const response = await fetch('/api/v1/orders/cities');

  if (!response.ok) {
    throw new OrdersApiError(response.status, await parseErrorMessage(response));
  }

  const data: unknown = await response.json();
  const cities = parseCities(data);

  if (cities.length === 0) {
    throw new OrdersApiError(500, 'Список городов пуст');
  }

  return cities;
};
