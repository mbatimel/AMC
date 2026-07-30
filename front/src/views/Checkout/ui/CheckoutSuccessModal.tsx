'use client';

import { Alert, Button, Description, Modal, useOverlayState } from '@heroui/react';
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

  return (
    <Modal state={state}>
      <Modal.Backdrop>
        <Modal.Container className={clsx(styles.successDialog)} size="md">
          <Modal.Dialog>
            <Modal.Header>
              <Modal.Heading>Заказ успешно оформлен</Modal.Heading>
              <Modal.CloseTrigger aria-label="Закрыть" />
            </Modal.Header>
            <Modal.Body className={clsx(styles.successBody)}>
              <Alert status="success">
                <Alert.Indicator />
                <Alert.Content>
                  <Alert.Description>
                    Заказ <strong>{order?.number ?? '—'}</strong> создан и передан в обработку
                  </Alert.Description>
                </Alert.Content>
              </Alert>

              {order ? (
                <dl className={clsx(styles.successMeta)}>
                  <div>
                    <dt>Сумма</dt>
                    <dd>{formatPrice(order.total)}</dd>
                  </div>
                  {order.deliveryType ? (
                    <div>
                      <dt>Доставка</dt>
                      <dd>{order.deliveryType}</dd>
                    </div>
                  ) : null}
                </dl>
              ) : null}

              <Description>Счёт и документы появятся в личном кабинете и на email.</Description>
            </Modal.Body>
            <Modal.Footer className={clsx(styles.successFooter)}>
              <Button
                onPress={() => {
                  onClose();
                  router.push(AppPath.CabinetOrders);
                }}
                variant="outline"
              >
                Перейти к заказам
              </Button>
              <Button
                onPress={() => {
                  onClose();
                  router.push(AppPath.Catalog);
                }}
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
