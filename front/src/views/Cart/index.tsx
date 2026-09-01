'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

import { useCart } from '@/core/entities/cart';
import { useCity } from '@/core/entities/city';
import { $productImageById } from '@/core/entities/productImages';
import { useSession } from '@/core/entities/session';
import { IconClose } from '@/core/shared/icons/IconClose';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { formatPositionsCount } from '@/core/shared/lib/pluralize';
import { AppPath, getProductPath } from '@/core/shared/router/paths';
import { Page } from '@/core/shared/ui/Page';
import { ProductThumb } from '@/core/shared/ui/ProductThumb';
import { QuantityStepper } from '@/core/shared/ui/QuantityStepper';
import { toastShown } from '@/core/shared/ui/Toast/model';

import styles from './Cart.module.css';

const downloadCartCsv = (
  items: {
    price: number;
    product_name: string;
    qty: number;
    sku: string;
    total: number;
  }[],
): void => {
  const rows = [
    ['Артикул', 'Наименование', 'Кол-во', 'Цена', 'Сумма'].join(';'),
    ...items.map((item) =>
      [item.sku, item.product_name, item.qty, item.price, item.total].join(';'),
    ),
  ];
  const blob = new Blob([`\uFEFF${rows.join('\n')}`], { type: 'text/csv;charset=utf-8;' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');

  link.href = url;
  link.download = 'cart.csv';
  link.click();
  URL.revokeObjectURL(url);
};

const IconDownload = (): JSX.Element => (
  <svg aria-hidden fill="none" height={14} viewBox="0 0 16 16" width={14}>
    <path
      d="M8 2v8m0 0 3-3M8 10 5 7M3 13h10"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
  </svg>
);

const IconUpload = (): JSX.Element => (
  <svg aria-hidden fill="none" height={14} viewBox="0 0 16 16" width={14}>
    <path
      d="M8 14V6m0 0 3 3M8 6 5 9M3 3h10"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
  </svg>
);

const IconTrash = (): JSX.Element => (
  <svg aria-hidden fill="none" height={14} viewBox="0 0 16 16" width={14}>
    <path
      d="M3 4h10M6 4V3h4v1m-5 2v6m3-6v6M4 4l.5 9h7L12 4"
      stroke="currentColor"
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth="1.5"
    />
  </svg>
);

export const CartPage = (): JSX.Element => {
  const router = useRouter();
  const showToast = useUnit(toastShown);
  const { isAuthenticated, isHydrated } = useSession();
  const { selectedCityName } = useCity();
  const { cart, cartCount, changeItemQty, clear, isCartPending, removeItem } = useCart();
  const productImages = useUnit($productImageById);

  useEffect(() => {
    if (!isHydrated) {
      return;
    }

    if (!isAuthenticated) {
      router.replace(`${AppPath.Login}?next=${encodeURIComponent(AppPath.Cart)}`);
    }
  }, [isAuthenticated, isHydrated, router]);

  if (!isHydrated) {
    return (
      <Page>
        <div className={clsx(styles.root)}>
          <div className={clsx(styles.container)}>
            <p className={clsx(styles.status)}>Загрузка…</p>
          </div>
        </div>
      </Page>
    );
  }

  if (!isAuthenticated) {
    return (
      <Page>
        <div className={clsx(styles.root)}>
          <div className={clsx(styles.container)}>
            <p className={clsx(styles.status)}>Требуется авторизация…</p>
          </div>
        </div>
      </Page>
    );
  }

  const isEmpty = cart.items.length === 0;
  const deliveryCity = selectedCityName || 'Самара';

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <div className={clsx(styles.container)}>
          <nav aria-label="Хлебные крошки" className={clsx(styles.breadcrumbs)}>
            <Link href={AppPath.Home}>Главная</Link>
            <span>/</span>
            <span>Корзина</span>
          </nav>

          <div className={clsx(styles.header)}>
            <h1 className={clsx(styles.title)}>Корзина · {formatPositionsCount(cartCount)}</h1>
            <div className={clsx(styles.toolbar)}>
              <button
                className={clsx(styles.toolButton)}
                disabled={isEmpty}
                onClick={() => downloadCartCsv(cart.items)}
                type="button"
              >
                {!isEmpty ? <IconDownload /> : null}
                Скачать CSV
              </button>
              {!isEmpty ? (
                <button
                  className={clsx(styles.toolButton)}
                  onClick={() => showToast({ message: 'Импорт CSV/Excel скоро будет доступен' })}
                  type="button"
                >
                  <IconUpload />
                  Импорт CSV/Excel
                </button>
              ) : null}
              <button
                className={clsx(styles.toolButton)}
                disabled={isEmpty || isCartPending}
                onClick={() => clear()}
                type="button"
              >
                {!isEmpty ? <IconTrash /> : null}
                Очистить
              </button>
            </div>
          </div>

          {isEmpty ? (
            <div className={clsx(styles.empty)}>
              <h2 className={clsx(styles.emptyTitle)}>Корзина пуста</h2>
              <p className={clsx(styles.emptyText)}>Добавьте товары из каталога</p>
              <Link className={clsx(styles.catalogLink)} href={AppPath.Catalog}>
                Перейти в каталог
              </Link>
            </div>
          ) : (
            <div className={clsx(styles.layout)}>
              <div className={clsx(styles.tableWrap)}>
                <table className={clsx(styles.table)}>
                  <thead>
                    <tr>
                      <th>Товар</th>
                      <th>Цена</th>
                      <th>Кол-во</th>
                      <th>Сумма</th>
                      <th aria-label="Удалить" />
                    </tr>
                  </thead>
                  <tbody>
                    {cart.items.map((item) => (
                      <tr key={item.id}>
                        <td>
                          <div className={clsx(styles.productCell)}>
                            <div className={clsx(styles.thumb)}>
                              <ProductThumb
                                alt={item.product_name}
                                fallbackClassName={clsx(styles.thumbFallback)}
                                src={productImages[item.product_id]}
                              />
                            </div>
                            <div className={clsx(styles.productInfo)}>
                              <Link
                                className={clsx(styles.productLink)}
                                href={getProductPath(item.product_id)}
                              >
                                {item.product_name}
                              </Link>
                              <div className={clsx(styles.meta)}>Арт. {item.sku}</div>
                            </div>
                          </div>
                        </td>
                        <td>
                          <div className={clsx(styles.priceCell)}>
                            <span className={clsx(styles.priceValue)}>
                              {formatPrice(item.price)}
                            </span>
                            <span className={clsx(styles.priceUnit)}>за шт</span>
                          </div>
                        </td>
                        <td>
                          <QuantityStepper
                            disabled={isCartPending}
                            onChange={(qty) => {
                              if (qty === item.qty) {
                                return;
                              }

                              changeItemQty({ cartItemID: item.id, qty });
                            }}
                            value={item.qty}
                          />
                        </td>
                        <td className={clsx(styles.sum)}>{formatPrice(item.total)}</td>
                        <td className={clsx(styles.removeCell)}>
                          <button
                            aria-label="Удалить"
                            className={clsx(styles.remove)}
                            disabled={isCartPending}
                            onClick={() => removeItem(item.id)}
                            type="button"
                          >
                            <IconClose height={16} width={16} />
                          </button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>

              <aside className={clsx(styles.summary)}>
                <h2 className={clsx(styles.summaryTitle)}>Итого по заказу</h2>
                <div className={clsx(styles.summaryRow)}>
                  <span>Сумма позиций</span>
                  <strong>{formatPrice(cart.subtotal || cart.total)}</strong>
                </div>
                {cart.discount_total > 0 ? (
                  <div className={clsx(styles.summaryRow, styles.discount)}>
                    <span>Скидка от объёма</span>
                    <strong>−{formatPrice(cart.discount_total)}</strong>
                  </div>
                ) : null}
                {cart.vat > 0 ? (
                  <div className={clsx(styles.summaryRow)}>
                    <span>в т.ч. НДС 20%</span>
                    <strong>{formatPrice(cart.vat)}</strong>
                  </div>
                ) : null}
                <div className={clsx(styles.totalRow)}>
                  <span>К оплате</span>
                  <strong>{formatPrice(cart.total)}</strong>
                </div>

                <div className={clsx(styles.tips)}>
                  <div className={clsx(styles.tip, styles.tipSuccess)}>
                    Бесплатная доставка · {deliveryCity}
                  </div>
                  <div className={clsx(styles.tip, styles.tipInfo)}>
                    Резерв по счёту: 5 банковских дней
                  </div>
                </div>

                <Button
                  className={clsx(styles.checkoutButton)}
                  fullWidth
                  onPress={() => router.push(AppPath.Checkout)}
                  variant="primary"
                >
                  Оформить заказ
                </Button>
                <Link className={clsx(styles.continueLink)} href={AppPath.Catalog}>
                  Продолжить покупки
                </Link>
              </aside>
            </div>
          )}
        </div>
      </div>
    </Page>
  );
};
