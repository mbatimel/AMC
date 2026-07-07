import { HttpClient } from './HttpClient';

export class Manager {
  protected readonly httpClient: HttpClient;

  constructor(httpClient: HttpClient = new HttpClient()) {
    this.httpClient = httpClient;
  }
}
