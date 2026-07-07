export type ApiAdditionalErrors = {
  errors?: ApiValidationError[] | null;
  reason?: string;
};

export type ApiEnvelope<TData = unknown> = {
  additionalErrors?: ApiAdditionalErrors;
  data: TData;
  error: boolean;
  errorText?: string;
};

export type ApiValidationError = {
  params?: Record<string, string>;
  trKey: string;
};
