import { apiUrl } from './baseUrl';
import { buildQueryString, type QueryParams } from './query';

export type HttpRequestOptions = {
  headers?: Record<string, string>;
  query?: QueryParams;
};

export type HttpResult<TBody = unknown> = {
  body: TBody;
  status: number;
};

export class HttpClient {
  async post<TBody = unknown>(
    path: string,
    options: HttpRequestOptions = {},
  ): Promise<HttpResult<TBody>> {
    const queryString = buildQueryString(options.query ?? {});
    const url = `${apiUrl(path)}${queryString}`;

    const response = await fetch(url, {
      credentials: 'include',
      headers: options.headers,
      method: 'POST',
    });

    const body = (await response.json()) as TBody;

    return {
      body,
      status: response.status,
    };
  }
}
