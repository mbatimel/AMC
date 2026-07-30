'use client';

import { Alert, Description, Link as HeroLink, Surface, Typography } from '@heroui/react';
import clsx from 'clsx';

import type { Cart } from '@/core/shared/api/cart';

import { formatPrice } from '@/core/shared/lib/formatPrice';
import { formatPositionsCount } from '@/core/shared/lib/pluralize';
import { AppPath, getProductPath } from '@/core/shared/router/paths';

import styles from '../Checkout.module.css';

type CheckoutSummaryProps = {
  cart: Cart;
  cityName: string;
};

export const CheckoutSummary = ({ cart, cityName }: CheckoutSummaryProps): JSX.Element => {
  return (
    <Surface className={clsx(styles.summary)}>
      <Typography.Heading className={clsx(styles.summaryTitle)} level={2}>
        Ваш заказ
      </Typography.Heading>
      <Description className={clsx(styles.summaryMeta)}>
        {formatPositionsCount(cart.items.length)}
      </Description>

      <ul className={clsx(styles.items)}>
        {cart.items.map((item) => (
          <li className={clsx(styles.item)} key={item.id}>
            <div>
              <HeroLink className={clsx(styles.itemName)} href={getProductPath(item.product_id)}>
                {item.product_name}
              </HeroLink>
              <Description>
                {item.sku ? `${item.sku} · ` : ''}
                {item.qty} шт.
              </Description>
            </div>
            <Typography className={clsx(styles.itemTotal)} weight="bold">
              {formatPrice(item.total)}
            </Typography>
          </li>
        ))}
      </ul>

      <div className={clsx(styles.summaryRow)}>
        <span>Сумма позиций</span>
        <strong>{formatPrice(cart.subtotal || cart.total)}</strong>
      </div>
      {cart.discount_total > 0 ? (
        <div className={clsx(styles.summaryRow, styles.discount)}>
          <span>Скидка</span>
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
        <Alert status="success">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Description>Бесплатная доставка · {cityName}</Alert.Description>
          </Alert.Content>
        </Alert>
        <Alert status="accent">
          <Alert.Indicator />
          <Alert.Content>
            <Alert.Description>Резерв по счёту: 5 банковских дней</Alert.Description>
          </Alert.Content>
        </Alert>
      </div>

      <HeroLink className={clsx(styles.backLink)} href={AppPath.Cart}>
        ← Вернуться в корзину
      </HeroLink>
    </Surface>
  );
};
