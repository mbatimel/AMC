import { createEffect, createEvent, createStore, sample } from 'effector';

import type { ProductListItem } from '@/core/shared/api/products';

import { askAssistantRequest } from '@/core/shared/api/assistant';
import { listProductsRequest } from '@/core/shared/api/products';
import { createSupportRequest } from '@/core/shared/api/support';

import type { AssistantMessage } from './lib/types';

import {
  ASSISTANT_FALLBACK_REPLY,
  ASSISTANT_GREETING,
  findAssistantScript,
} from './lib/scripts';

const STORAGE_KEY = 'amc_assistant_chat';
const OPERATOR_WAIT_MS = 2500;
const MAX_SUGGESTED_PRODUCTS = 3;

const createMessageId = (): string =>
  `msg-${Date.now().toString(36)}-${Math.random().toString(36).slice(2, 7)}`;

const nowTime = (): string =>
  new Date().toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });

const greetingMessage = (): AssistantMessage => ({
  author: 'bot',
  id: 'greeting',
  text: ASSISTANT_GREETING,
  time: nowTime(),
});

const readHistory = (): AssistantMessage[] => {
  if (typeof window === 'undefined') {
    return [greetingMessage()];
  }

  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);

    if (!raw) {
      return [greetingMessage()];
    }

    const parsed: unknown = JSON.parse(raw);

    if (!Array.isArray(parsed) || parsed.length === 0) {
      return [greetingMessage()];
    }

    return parsed as AssistantMessage[];
  } catch {
    return [greetingMessage()];
  }
};

export const assistantHydrated = createEvent();
export const assistantOpened = createEvent();
export const assistantClosed = createEvent();
export const assistantToggled = createEvent();
export const assistantReset = createEvent();
export const assistantMessageSent = createEvent<string>();
export const assistantOperatorRequested = createEvent<{
  clientName?: string;
  contact?: string;
  userId?: string;
}>();

const searchProducts = async (query: string): Promise<ProductListItem[]> => {
  try {
    const result = await listProductsRequest({ limit: MAX_SUGGESTED_PRODUCTS, q: query });

    return result.items.slice(0, MAX_SUGGESTED_PRODUCTS);
  } catch {
    return [];
  }
};

/**
 * Встроенный ответ без LLM: поиск по каталогу + сценарии предметной области.
 * Используется, когда DeepSeek не подключён или недоступен.
 */
const answerLocally = async (text: string): Promise<AssistantMessage> => {
  const script = findAssistantScript(text);
  let products = await searchProducts(script?.query || text);

  if (products.length === 0 && script?.query) {
    products = await searchProducts(text);
  }

  if (script) {
    return {
      author: 'bot',
      id: createMessageId(),
      offerOperator: products.length === 0,
      products,
      text: script.reply,
      time: nowTime(),
    };
  }

  if (products.length > 0) {
    return {
      author: 'bot',
      id: createMessageId(),
      products,
      text: `Нашёл подходящие позиции по запросу «${text}». Проверьте артикул, ГОСТ и наличие.`,
      time: nowTime(),
    };
  }

  return {
    author: 'bot',
    id: createMessageId(),
    offerOperator: true,
    products: [],
    text: ASSISTANT_FALLBACK_REPLY,
    time: nowTime(),
  };
};

export type AssistantAnswerParams = {
  history: AssistantMessage[];
  text: string;
};

/**
 * Ответ помощника. Сначала пробуем ИИ (`/portal-api/assistant` → DeepSeek
 * с поиском по каталогу через tool calls). Если ключ не задан, модель
 * недоступна или вернула пустой ответ — молча уходим на встроенную логику.
 */
export const assistantAnswerFx = createEffect(
  async ({ history, text }: AssistantAnswerParams): Promise<AssistantMessage> => {
    try {
      const remote = await askAssistantRequest({
        history: history
          .filter((message) => message.author === 'bot' || message.author === 'user')
          .filter((message) => !(message.author === 'user' && message.text === text))
          .slice(-8)
          .map((message) => ({ author: message.author, text: message.text })),
        message: text,
      });

      if (remote.configured && remote.reply) {
        return {
          author: 'bot',
          id: createMessageId(),
          offerOperator: remote.offerOperator ?? false,
          products: remote.products ?? [],
          text: remote.reply,
          time: nowTime(),
        };
      }
    } catch {
      // сеть или 5xx — работаем на встроенной логике
    }

    return answerLocally(text);
  },
);

export const assistantOperatorFx = createEffect(
  async (payload: {
    clientName?: string;
    contact?: string;
    history: AssistantMessage[];
    userId?: string;
  }): Promise<void> => {
    const transcript = payload.history
      .slice(-12)
      .map((message) => `${message.author}: ${message.text}`)
      .join('\n');

    await createSupportRequest({
      clientName: payload.clientName,
      contact: payload.contact,
      severity: 3,
      source: 'assistant',
      subject: 'Запрос оператора из чата подбора инструмента',
      text: transcript || 'Клиент запросил оператора без истории диалога.',
      userId: payload.userId,
    });

    await new Promise((resolve) => {
      setTimeout(resolve, OPERATOR_WAIT_MS);
    });
  },
);

const persistHistoryFx = createEffect((history: AssistantMessage[]) => {
  try {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(history.slice(-40)));
  } catch {
    // приватный режим браузера — история живёт только в памяти
  }
});

export const $isAssistantOpen = createStore(false)
  .on(assistantOpened, () => true)
  .on(assistantClosed, () => false)
  .on(assistantToggled, (isOpen) => !isOpen);

export const $assistantHistory = createStore<AssistantMessage[]>([greetingMessage()])
  .on(assistantHydrated, () => readHistory())
  .on(assistantMessageSent, (history, text) => [
    ...history,
    { author: 'user' as const, id: createMessageId(), text, time: nowTime() },
  ])
  .on(assistantAnswerFx.doneData, (history, message) => [...history, message])
  .on(assistantAnswerFx.fail, (history) => [
    ...history,
    {
      author: 'system' as const,
      id: createMessageId(),
      text: 'Не удалось получить ответ. Попробуйте ещё раз или позовите оператора.',
      time: nowTime(),
    },
  ])
  .on(assistantOperatorRequested, (history) => [
    ...history,
    {
      author: 'system' as const,
      id: createMessageId(),
      text: 'Передаём диалог оператору. Обращение зарегистрировано в поддержке.',
      time: nowTime(),
    },
  ])
  .on(assistantOperatorFx.done, (history) => [
    ...history,
    {
      author: 'operator' as const,
      id: createMessageId(),
      text: 'Здравствуйте! Вижу вашу переписку с помощником. Напишите, что нужно уточнить — артикул, размер или номер заказа.',
      time: nowTime(),
    },
  ])
  .on(assistantOperatorFx.fail, (history) => [
    ...history,
    {
      author: 'system' as const,
      id: createMessageId(),
      text: 'Не удалось связаться с оператором. Напишите в раздел «Поддержка» — обращение попадёт менеджеру.',
      time: nowTime(),
    },
  ])
  .reset(assistantReset);

export const $isAssistantPending = assistantAnswerFx.pending;

export const $isOperatorPending = assistantOperatorFx.pending;

export const $isOperatorMode = createStore(false)
  .on(assistantOperatorFx.done, () => true)
  .reset(assistantReset);

sample({
  clock: assistantMessageSent,
  source: $assistantHistory,
  filter: (_, text) => text.trim().length > 0,
  fn: (history, text) => ({ history, text: text.trim() }),
  target: assistantAnswerFx,
});

sample({
  clock: assistantOperatorRequested,
  source: $assistantHistory,
  fn: (history, payload) => ({ ...payload, history }),
  target: assistantOperatorFx,
});

sample({
  clock: $assistantHistory,
  target: persistHistoryFx,
});
