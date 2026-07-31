'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';

import { ASSISTANT_SUGGESTIONS, useAssistant } from '@/core/entities/assistant';
import { InfoCard, InfoPage } from '@/core/shared/ui/InfoPage';
import { Page } from '@/core/shared/ui/Page';

import styles from './Assistant.module.css';

export const Assistant = (): JSX.Element => {
  const { open, reset, send } = useAssistant();

  const ask = (question: string): void => {
    open();
    send(question);
  };

  return (
    <Page>
      <InfoPage
        description="Опишите задачу — помощник подберёт позиции по каталогу, ГОСТу и размеру, а при необходимости подключит оператора."
        eyebrow="Помощник"
        title="Подбор инструмента"
      >
        <InfoCard title="Как это работает">
          <ol className={clsx(styles.steps)}>
            <li>Опишите операцию, материал и размер — либо введите артикул или ГОСТ.</li>
            <li>Помощник ищет позиции в каталоге и показывает цену и остаток.</li>
            <li>Товар можно сразу открыть или добавить в корзину.</li>
            <li>
              Если подходящего не нашлось — кнопка «Позвать оператора» создаёт обращение в
              поддержке вместе с историей диалога.
            </li>
          </ol>
        </InfoCard>

        <InfoCard title="Частые запросы">
          <div className={clsx(styles.suggestions)}>
            {ASSISTANT_SUGGESTIONS.map((suggestion) => (
              <button
                className={clsx(styles.suggestion)}
                key={suggestion}
                onClick={() => ask(suggestion)}
                type="button"
              >
                {suggestion}
              </button>
            ))}
          </div>
        </InfoCard>

        <InfoCard>
          <div className={clsx(styles.actions)}>
            <Button onPress={() => open()} variant="primary">
              Открыть помощника
            </Button>
            <Button onPress={() => reset()} variant="outline">
              Очистить историю диалога
            </Button>
          </div>
        </InfoCard>
      </InfoPage>
    </Page>
  );
};
