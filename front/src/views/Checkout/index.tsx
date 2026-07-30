'use client';

import { Breadcrumbs, Description, Spinner, Typography } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';

import { useCart } from '@/core/entities/cart';
import { useCity } from '@/core/entities/city';
import { useSession } from '@/core/entities/session';
import { AppPath } from '@/core/shared/router/paths';
import { Page } from '@/core/shared/ui/Page';
import { ToastViewport } from '@/core/shared/ui/Toast';

import styles from './Checkout.module.css';
import {
  $checkoutError,
  $isCheckoutPending,
  $isSuccessOpen,
  $successOrderView,
  checkoutSubmitted,
  checkoutSuccessClosed,
} from './model';
import { CheckoutForm } from './ui/CheckoutForm';
import { CheckoutSuccessModal } from './ui/CheckoutSuccessModal';
import { CheckoutSummary } from './ui/CheckoutSummary';

export const CheckoutPage = (): JSX.Element => {
  const router = useRouter();
  const { isAuthenticated, isHydrated } = useSession();
  const { selectedCityName } = useCity();
  const { cart, cartCount, isCartPending } = useCart();
  const [error, isPending, isSuccessOpen, successOrder, submit, closeSuccess] = useUnit([
    $checkoutError,
    $isCheckoutPending,
    $isSuccessOpen,
    $successOrderView,
    checkoutSubmitted,
    checkoutSuccessClosed,
  ]);

  useEffect(() => {
    if (!isHydrated) {
      return;
    }

    if (!isAuthenticated) {
      router.replace(`${AppPath.Login}?next=${encodeURIComponent(AppPath.Checkout)}`);
    }
  }, [isAuthenticated, isHydrated, router]);

  useEffect(() => {
    if (!isHydrated || !isAuthenticated || isSuccessOpen || isCartPending) {
      return;
    }

    if (cartCount === 0) {
      router.replace(AppPath.Cart);
    }
  }, [cartCount, isAuthenticated, isCartPending, isHydrated, isSuccessOpen, router]);

  if (!isHydrated || !isAuthenticated) {
    return (
      <Page>
        <div className={clsx(styles.root)}>
          <div className={clsx(styles.container, styles.statusRow)}>
            <Spinner size="sm" />
            <Description>Загрузка…</Description>
          </div>
        </div>
      </Page>
    );
  }

  if (!isSuccessOpen && (isCartPending || cartCount === 0)) {
    return (
      <Page>
        <div className={clsx(styles.root)}>
          <div className={clsx(styles.container, styles.statusRow)}>
            <Spinner size="sm" />
            <Description>
              {isCartPending ? 'Загрузка корзины…' : 'Корзина пуста, возвращаем…'}
            </Description>
          </div>
        </div>
      </Page>
    );
  }

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <div className={clsx(styles.container)}>
          <Breadcrumbs className={clsx(styles.breadcrumbs)}>
            <Breadcrumbs.Item href={AppPath.Home}>Главная</Breadcrumbs.Item>
            <Breadcrumbs.Item href={AppPath.Cart}>Корзина</Breadcrumbs.Item>
            <Breadcrumbs.Item>Оформление</Breadcrumbs.Item>
          </Breadcrumbs>

          <header className={clsx(styles.header)}>
            <Typography.Heading className={clsx(styles.title)} level={1}>
              Оформление заказа
            </Typography.Heading>
          </header>

          <div className={clsx(styles.layout)}>
            <CheckoutForm error={error} isPending={isPending} onSubmit={submit} />
            <CheckoutSummary cart={cart} cityName={selectedCityName || 'ваш город'} />
          </div>
        </div>

        <CheckoutSuccessModal isOpen={isSuccessOpen} onClose={closeSuccess} order={successOrder} />
        <ToastViewport />
      </div>
    </Page>
  );
};
