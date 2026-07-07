import type { ApiAdditionalErrors, ApiEnvelope } from './types';

export class ApiError extends Error {
  readonly additionalErrors?: ApiAdditionalErrors;
  readonly statusCode?: number;

  constructor(message: string, additionalErrors?: ApiAdditionalErrors, statusCode?: number) {
    super(message);
    this.name = 'ApiError';
    this.additionalErrors = additionalErrors;
    this.statusCode = statusCode;
  }

  static fromEnvelope(envelope: ApiEnvelope<unknown>, statusCode?: number): ApiError {
    return new ApiError(
      envelope.errorText ?? 'Unknown API error',
      envelope.additionalErrors,
      statusCode,
    );
  }
}
