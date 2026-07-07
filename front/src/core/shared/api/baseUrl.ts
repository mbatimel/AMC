import { env } from '@/core/shared/config/env';

export const apiUrl = (path: string): string => `${env.apiUrl}${path}`;
