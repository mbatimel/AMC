'use client';

import { CabinetOrder } from './ui/CabinetOrder';

type CabinetOrderPageProps = {
  orderId: string;
};

export const CabinetOrderPage = ({ orderId }: CabinetOrderPageProps): JSX.Element => {
  return <CabinetOrder orderId={orderId} />;
};
