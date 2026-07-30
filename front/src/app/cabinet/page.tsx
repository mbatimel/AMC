import { redirect } from 'next/navigation';

import { AppPath } from '@/core/shared/router/paths';

const Page = (): never => {
  redirect(AppPath.CabinetOrders);
};

export default Page;
