import type { AssistantConfig } from './config';

export type ChatMessage = {
  content: null | string;
  role: 'assistant' | 'system' | 'tool' | 'user';
  tool_call_id?: string;
  tool_calls?: ChatToolCall[];
};

export type ChatToolCall = {
  function: { arguments: string; name: string };
  id: string;
  type: string;
};

export type ChatCompletion = {
  finishReason: string;
  message: ChatMessage;
};

export class DeepseekError extends Error {
  readonly status: number;

  constructor(status: number, message: string) {
    super(message);
    this.name = 'DeepseekError';
    this.status = status;
  }
}

/**
 * Вызов chat-completions в формате, совместимом с OpenAI.
 * DeepSeek принимает тот же контракт: https://api-docs.deepseek.com/
 */
export const createChatCompletion = async ({
  config,
  messages,
  tools,
}: {
  config: AssistantConfig;
  messages: ChatMessage[];
  tools?: unknown[];
}): Promise<ChatCompletion> => {
  const response = await fetch(`${config.baseUrl}/chat/completions`, {
    body: JSON.stringify({
      max_tokens: 700,
      messages,
      model: config.model,
      stream: false,
      temperature: 0.2,
      ...(tools ? { tool_choice: 'auto', tools } : {}),
    }),
    headers: {
      Authorization: `Bearer ${config.apiKey}`,
      'Content-Type': 'application/json',
    },
    method: 'POST',
    signal: AbortSignal.timeout(config.timeoutMs),
  });

  if (!response.ok) {
    const detail = await response.text().catch(() => '');

    throw new DeepseekError(
      response.status,
      `DeepSeek вернул ${response.status}${detail ? `: ${detail.slice(0, 300)}` : ''}`,
    );
  }

  const payload: unknown = await response.json();

  if (typeof payload !== 'object' || payload === null) {
    throw new DeepseekError(502, 'Некорректный ответ модели');
  }

  const choices = (payload as Record<string, unknown>).choices;

  if (!Array.isArray(choices) || choices.length === 0) {
    throw new DeepseekError(502, 'Модель не вернула ответ');
  }

  const choice = choices[0] as Record<string, unknown>;
  const message = choice.message;

  if (typeof message !== 'object' || message === null) {
    throw new DeepseekError(502, 'Модель не вернула сообщение');
  }

  const messageRecord = message as Record<string, unknown>;

  return {
    finishReason: typeof choice.finish_reason === 'string' ? choice.finish_reason : '',
    message: {
      content: typeof messageRecord.content === 'string' ? messageRecord.content : null,
      role: 'assistant',
      tool_calls: Array.isArray(messageRecord.tool_calls)
        ? (messageRecord.tool_calls as ChatToolCall[])
        : undefined,
    },
  };
};
