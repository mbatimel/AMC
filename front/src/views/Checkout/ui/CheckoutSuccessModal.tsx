'use client';

import { Button, Modal, useOverlayState } from '@heroui/react';
import clsx from 'clsx';
import { useRouter } from 'next/navigation';

import { formatPrice } from '@/core/shared/lib/formatPrice';
import { AppPath } from '@/core/shared/router/paths';

import styles from '../Checkout.module.css';

export type CheckoutSuccessOrder = {
  deliveryType?: string;
  number: string;
  total: number;
};

type CheckoutSuccessModalProps = {
  isOpen: boolean;
  onClose: () => void;
  order: CheckoutSuccessOrder | null;
};

export const CheckoutSuccessModal = ({
  isOpen,
  onClose,
  order,
}: CheckoutSuccessModalProps): JSX.Element => {
  const router = useRouter();
  const state = useOverlayState({
    isOpen,
    onOpenChange: (open) => {
      if (!open) {
        onClose();
      }
    },
  });

  /** Не закрываем модалку до ухода со страницы — иначе пустая корзина редиректит на /cart. */
  const leaveTo = (path: string): void => {
    router.push(path);
  };

  return (
    <Modal state={state}>
      <Modal.Backdrop>
        <Modal.Container className={clsx(styles.successDialog)} size="lg">
          <Modal.Dialog>
            <Modal.Header>
              <Modal.Heading>Заказ успешно оформлен</Modal.Heading>
              <Modal.CloseTrigger aria-label="Закрыть" />
            </Modal.Header>
            <Modal.Body className={clsx(styles.successBody)}>
              <p className={clsx(styles.successBanner)} role="status">
                Заказ <strong>{order?.number ?? '—'}</strong> создан и передан в обработку
              </p>

              {order ? (
                <dl className={clsx(styles.successMeta)}>
                  <div>
                    <dt>Сумма</dt>
                    <dd>{formatPrice(order.total)}</dd>
                  </div>
                  {order.deliveryType ? (
                    <div>
                      <dt>Способ доставки</dt>
                      <dd>{order.deliveryType}</dd>
                    </div>
                  ) : null}
                </dl>
              ) : null}

              <p className={clsx(styles.successInfo)}>
                После подтверждения в течение нескольких минут вы получите счёт в ЛК и по e-mail.
              </p>
            </Modal.Body>
            <Modal.Footer className={clsx(styles.successFooter)}>
              <Button
                className={clsx(styles.successSecondary)}
                onPress={() => leaveTo(AppPath.CabinetOrders)}
                variant="outline"
              >
                Перейти к заказам
              </Button>
              <Button
                className={clsx(styles.successPrimary)}
                onPress={() => leaveTo(AppPath.Catalog)}
                variant="primary"
              >
                Продолжить покупки
              </Button>
            </Modal.Footer>
          </Modal.Dialog>
        </Modal.Container>
      </Modal.Backdrop>
    </Modal>
  );
};
