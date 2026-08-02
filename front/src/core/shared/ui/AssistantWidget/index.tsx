'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useRouter } from 'next/navigation';
import { useEffect, useRef, useState } from 'react';

import { ASSISTANT_SUGGESTIONS, useAssistant } from '@/core/entities/assistant';
import { useCart } from '@/core/entities/cart';
import { useSession } from '@/core/entities/session';
import { IconClose, IconSearch, IconSupport } from '@/core/shared/icons';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { getPrimaryProductImageUrl } from '@/core/shared/lib/productImage';
import { getProductPath } from '@/core/shared/router/paths';
import { ProductThumb } from '@/core/shared/ui/ProductThumb';

import styles from './AssistantWidget.module.css';

/**
 * Плавающий помощник по подбору инструмента (M-04).
 * Ищет позиции в каталоге, отвечает по сценариям и умеет позвать оператора —
 * запрос оператора создаёт обращение в поддержке.
 */
export const AssistantWidget = (): JSX.Element => {
  const router = useRouter();
  const { addToCart } = useCart();
  const { userId } = useSession();
  const {
    close,
    history,
    isOpen,
    isOperatorMode,
    isOperatorPending,
    isPending,
    requestOperator,
    send,
    toggle,
  } = useAssistant();
  const [draft, setDraft] = useState('');
  const bodyRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (isOpen && bodyRef.current) {
      bodyRef.current.scrollTop = bodyRef.current.scrollHeight;
    }
  }, [history, isOpen]);

  const submit = (text: string): void => {
    const trimmed = text.trim();

    if (!trimmed || isPending) {
      return;
    }

    send(trimmed);
    setDraft('');
  };

  const showSuggestions = !isOperatorMode && !isOperatorPending && history.length <= 2;

  return (
    <>
      <button
        aria-label={isOpen ? 'Свернуть подбор инструмента' : 'Открыть подбор инструмента'}
        className={clsx(styles.floatButton, isOpen && styles.floatButtonHidden)}
        onClick={() => toggle()}
        type="button"
      >
        <IconSearch currentColor="currentColor" height={18} width={18} />
        <span className={clsx(styles.floatLabel)}>Подбор инструмента</span>
      </button>

      <section
        aria-hidden={!isOpen}
        aria-label="Подбор инструмента"
        className={clsx(styles.panel, isOpen && styles.panelOpen)}
      >
        <header className={clsx(styles.header)}>
          <div>
            <p className={clsx(styles.headerTitle)}>Подбор инструмента</p>
            <p className={clsx(styles.headerSubtitle)}>
              {isOperatorMode ? 'Оператор на связи' : 'Поиск по каталогу и подсказки'}
            </p>
          </div>
          <button
            aria-label="Закрыть"
            className={clsx(styles.closeButton)}
            onClick={() => close()}
            type="button"
          >
            <IconClose currentColor="currentColor" height={16} width={16} />
          </button>
        </header>

        <div className={clsx(styles.body)} ref={bodyRef}>
          {history.map((message) => {
            if (message.author === 'system') {
              return (
                <p className={clsx(styles.system)} key={message.id}>
                  {message.text}
                </p>
              );
            }

            return (
              <div
                className={clsx(
                  styles.message,
                  message.author === 'user' && styles.messageUser,
                  message.author === 'operator' && styles.messageOperator,
                )}
                key={message.id}
              >
                {message.author === 'operator' ? (
                  <span className={clsx(styles.messageAuthor)}>Оператор</span>
                ) : null}
                <p className={clsx(styles.messageText)}>{message.text}</p>

                {message.products && message.products.length > 0 ? (
                  <ul className={clsx(styles.products)}>
                    {message.products.map((product) => (
                      <li className={clsx(styles.product)} key={product.id}>
                        <div className={clsx(styles.productThumb)}>
                          <ProductThumb
                            alt={product.name}
                            categoryName={product.category_name}
                            images={product.images}
                          />
                        </div>
                        <div className={clsx(styles.productBody)}>
                          <p className={clsx(styles.productSku)}>
                            {product.sku}
                            {product.gost ? ` · ${product.gost}` : ''}
                          </p>
                          <p className={clsx(styles.productName)}>{product.name}</p>
                          <p className={clsx(styles.productPrice)}>
                            {formatPrice(product.client_price || product.base_price)}
                            {product.stock_qty > 0
                              ? ` · в наличии ${product.stock_qty}`
                              : ' · под заказ'}
                          </p>
                          <div className={clsx(styles.productActions)}>
                            <button
                              className={clsx(styles.productButton)}
                              onClick={() => {
                                close();
                                router.push(getProductPath(product.id));
                              }}
                              type="button"
                            >
                              Открыть
                            </button>
                            <button
                              className={clsx(styles.productButton, styles.productButtonPrimary)}
                              disabled={product.stock_qty <= 0}
                              onClick={() =>
                                addToCart({
                                  imageUrl: getPrimaryProductImageUrl(product.images) ?? undefined,
                                  name: product.name,
                                  productID: product.id,
                                  qty: product.package_qty,
                                })
                              }
                              type="button"
                            >
                              В корзину
                            </button>
                          </div>
                        </div>
                      </li>
                    ))}
                  </ul>
                ) : null}

                {message.offerOperator && !isOperatorMode && !isOperatorPending ? (
                  <button
                    className={clsx(styles.operatorInline)}
                    onClick={() => requestOperator({ userId: userId ?? undefined })}
                    type="button"
                  >
                    <IconSupport currentColor="currentColor" height={13} width={13} />
                    Позвать оператора
                  </button>
                ) : null}
              </div>
            );
          })}

          {isPending ? <p className={clsx(styles.typing)}>Подбираю варианты…</p> : null}
          {isOperatorPending ? (
            <p className={clsx(styles.typing)}>Ищем свободного оператора…</p>
          ) : null}
        </div>

        {showSuggestions ? (
          <div className={clsx(styles.suggestions)}>
            {ASSISTANT_SUGGESTIONS.map((suggestion) => (
              <button
                className={clsx(styles.suggestion)}
                key={suggestion}
                onClick={() => submit(suggestion)}
                type="button"
              >
                {suggestion}
              </button>
            ))}
          </div>
        ) : null}

        <form
          className={clsx(styles.form)}
          onSubmit={(event) => {
            event.preventDefault();
            submit(draft);
          }}
        >
          <input
            aria-label="Сообщение помощнику"
            className={clsx(styles.input)}
            onChange={(event) => setDraft(event.target.value)}
            placeholder="Артикул, ГОСТ, размер или задача"
            value={draft}
          />
          <Button
            isDisabled={isPending || draft.trim().length === 0}
            type="submit"
            variant="primary"
          >
            Отправить
          </Button>
        </form>

        {!isOperatorMode && !isOperatorPending ? (
          <button
            className={clsx(styles.operatorLink)}
            onClick={() => requestOperator({ userId: userId ?? undefined })}
            type="button"
          >
            Не нашли нужное? Позвать оператора
          </button>
        ) : null}
      </section>
    </>
  );
};
