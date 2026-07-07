const apiUrl = process.env.NEXT_PUBLIC_API_URL ?? '';

export const env = {
  apiUrl,
} as const;
