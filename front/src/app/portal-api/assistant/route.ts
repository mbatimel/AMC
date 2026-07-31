import type { ProductListItem } from '@/core/shared/api/products';
import type { ChatMessage } from '@/core/shared/server/assistant/deepseek';

import { apiFail, apiOk, readJsonBody } from '@/core/shared/server/portal/response';
import {
  searchCatalogProducts,
  toModelProduct,
} from '@/core/shared/server/assistant/catalog';
import { readAssistantConfig } from '@/core/shared/server/assistant/config';
import { createChatCompletion } from '@/core/shared/server/assistant/deepseek';
import { ASSISTANT_SYSTEM_PROMPT, CATALOG_TOOL } from '@/core/shared/server/assistant/prompt';

export const dynamic = 'force-dynamic';

/** Сколько раз подряд модель может дёрнуть search_catalog за один запрос. */
const MAX_TOOL_ROUNDS = 3;

type AssistantRequestBody = {
  history?: { author?: string; text?: string }[];
  message?: string;
};

type ToolArguments = {
  gost?: string;
  inStock?: boolean;
  material?: string;
  q?: string;
  size?: string;
};

const parseToolArguments = (raw: string): ToolArguments => {
  try {
    const parsed: unknown = JSON.parse(raw);

    return typeof parsed === 'object' && parsed !== null ? (parsed as ToolArguments) : {};
  } catch {
    return {};
  }
};

const toChatRole = (author: string | undefined): 'assistant' | 'user' =>
  author === 'user' ? 'user' : 'assistant';

export const POST = async (request: Request): Promise<Response> => {
  const config = readAssistantConfig();
  const body = await readJsonBody<AssistantRequestBody>(request);
  const message = body?.message?.trim();

  if (!message) {
    return apiFail(400, 'Пустой запрос');
  }

  /**
   * Ключ не задан — портал не считает это ошибкой: фронт молча уходит
   * на встроенный поиск по каталогу и сценарные ответы.
   */
  if (!config) {
    return apiOk({ configured: false });
  }

  const history = (body?.history ?? [])
    .filter((item) => typeof item.text === 'string' && item.text.trim().length > 0)
    .filter((item) => item.author === 'user' || item.author === 'bot')
    .slice(-config.maxHistory)
    .map<ChatMessage>((item) => ({
      content: item.text ?? '',
      role: toChatRole(item.author),
    }));

  const messages: ChatMessage[] = [
    { content: ASSISTANT_SYSTEM_PROMPT, role: 'system' },
    ...history,
    { content: message, role: 'user' },
  ];

  const foundProducts = new Map<string, ProductListItem>();

  try {
    for (let round = 0; round <= MAX_TOOL_ROUNDS; round += 1) {
      const isLastRound = round === MAX_TOOL_ROUNDS;
      const completion = await createChatCompletion({
        config,
        messages,
        tools: isLastRound ? undefined : [CATALOG_TOOL],
      });

      const toolCalls = completion.message.tool_calls ?? [];

      if (toolCalls.length === 0) {
        const reply = completion.message.content?.trim();

        if (!reply) {
          return apiOk({ configured: false });
        }

        return apiOk({
          configured: true,
          offerOperator: foundProducts.size === 0,
          products: [...foundProducts.values()].slice(0, config.maxProducts),
          reply,
        });
      }

      messages.push({
        content: completion.message.content,
        role: 'assistant',
        tool_calls: toolCalls,
      });

      for (const toolCall of toolCalls) {
        const args = parseToolArguments(toolCall.function.arguments);
        const products =
          toolCall.function.name === 'search_catalog'
            ? await searchCatalogProducts({
                gost: args.gost,
                inStock: args.inStock,
                limit: 5,
                material: args.material,
                q: args.q,
                size: args.size,
              })
            : [];

        products.forEach((product) => {
          foundProducts.set(product.id, product);
        });

        messages.push({
          content: JSON.stringify({
            items: products.map(toModelProduct),
            total: products.length,
          }),
          role: 'tool',
          tool_call_id: toolCall.id,
        });
      }
    }

    /** Модель зациклилась на поиске — отдаём то, что нашли. */
    return apiOk({
      configured: true,
      offerOperator: foundProducts.size === 0,
      products: [...foundProducts.values()].slice(0, config.maxProducts),
      reply:
        foundProducts.size > 0
          ? 'Подобрал позиции по вашему запросу — проверьте артикул, ГОСТ и наличие ниже.'
          : 'Не удалось однозначно подобрать позицию. Уточните артикул, ГОСТ или размер — либо подключите оператора.',
    });
  } catch (error) {
    /**
     * Модель недоступна, лимит запросов, таймаут — не роняем чат:
     * клиент по `configured: false` переходит на встроенную логику.
     */
    console.error('assistant: deepseek call failed', error);

    return apiOk({ configured: false });
  }
};
