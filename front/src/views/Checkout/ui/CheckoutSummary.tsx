'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';

import type { Cart } from '@/core/shared/api/cart';

import { $productImageById } from '@/core/entities/productImages';
import { formatPrice } from '@/core/shared/lib/formatPrice';
import { AppPath, getProductPath } from '@/core/shared/router/paths';
import { ProductThumb } from '@/core/shared/ui/ProductThumb';

import styles from '../Checkout.module.css';
import { CHECKOUT_FORM_ID } from './CheckoutForm';

type CheckoutSummaryProps = {
  cart: Cart;
  isPending: boolean;
};

export const CheckoutSummary = ({ cart, isPending }: CheckoutSummaryProps): JSX.Element => {
  const productImages = useUnit($productImageById);
  const subtotal = cart.subtotal || cart.total;
  const discountPercent =
    cart.discount_total > 0 && subtotal > 0
      ? Math.max(1, Math.round((cart.discount_total / subtotal) * 100))
      : 0;

  return (
    <aside className={clsx(styles.summary)}>
      <h2 className={clsx(styles.summaryTitle)}>Заказ</h2>

      <ul className={clsx(styles.items)}>
        {cart.items.map((item) => (
          <li className={clsx(styles.item)} key={item.id}>
            <div className={clsx(styles.itemThumb)}>
              <ProductThumb alt={item.product_name} src={productImages[item.product_id]} />
            </div>
            <div className={clsx(styles.itemMain)}>
              <Link className={clsx(styles.itemName)} href={getProductPath(item.product_id)}>
                {item.product_name}
                {item.sku ? ` ${item.sku}` : ''}
              </Link>
              <span className={clsx(styles.itemQty)}>{item.qty} шт</span>
            </div>
            <span className={clsx(styles.itemTotal)}>{formatPrice(item.total)}</span>
          </li>
        ))}
      </ul>

      <div className={clsx(styles.summaryRow)}>
        <span>Сумма</span>
        <strong>{formatPrice(subtotal)}</strong>
      </div>
      {cart.discount_total > 0 ? (
        <div className={clsx(styles.summaryRow, styles.discount)}>
          <span>{discountPercent > 0 ? `Скидка ${discountPercent}%` : 'Скидка'}</span>
          <strong>−{formatPrice(cart.discount_total)}</strong>
        </div>
      ) : null}
      <div className={clsx(styles.totalRow)}>
        <span>К оплате</span>
        <strong>{formatPrice(cart.total)}</strong>
      </div>

      <Button
        className={clsx(styles.submitButton)}
        form={CHECKOUT_FORM_ID}
        fullWidth
        isDisabled={isPending}
        type="submit"
        variant="primary"
      >
        {isPending ? 'Оформляем…' : '✓ Оформить заказ'}
      </Button>

      <Link className={clsx(styles.backLink)} href={AppPath.Cart}>
        Назад в корзину
      </Link>
    </aside>
  );
};
