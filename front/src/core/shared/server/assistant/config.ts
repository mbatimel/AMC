/**
 * Конфигурация ИИ-помощника. Ключ читается только на сервере —
 * переменная намеренно без префикса NEXT_PUBLIC_, чтобы не попасть в бандл.
 */

export type AssistantConfig = {
  apiKey: string;
  baseUrl: string;
  maxHistory: number;
  maxProducts: number;
  model: string;
  timeoutMs: number;
};

/**
 * `deepseek-chat` признан устаревшим 24.07.2026, поэтому по умолчанию
 * используем актуальную быструю модель. Переопределяется `DEEPSEEK_MODEL`.
 */
const DEFAULT_MODEL = 'deepseek-v4-flash';

const DEFAULT_BASE_URL = 'https://api.deepseek.com';

export const readAssistantConfig = (): AssistantConfig | null => {
  const apiKey = process.env.DEEPSEEK_API_KEY?.trim();

  if (!apiKey) {
    return null;
  }

  return {
    apiKey,
    baseUrl: (process.env.DEEPSEEK_BASE_URL?.trim() || DEFAULT_BASE_URL).replace(/\/$/, ''),
    maxHistory: Number(process.env.ASSISTANT_MAX_HISTORY) || 8,
    maxProducts: Number(process.env.ASSISTANT_MAX_PRODUCTS) || 3,
    model: process.env.DEEPSEEK_MODEL?.trim() || DEFAULT_MODEL,
    timeoutMs: Number(process.env.DEEPSEEK_TIMEOUT_MS) || 25_000,
  };
};

/** Базовый адрес backend-сервисов для серверных запросов к каталогу. */
export const readApiBaseUrl = (): string =>
  (process.env.API_PROXY_TARGET ?? 'https://wk.amctechgroup.ru').replace(/\/$/, '');
