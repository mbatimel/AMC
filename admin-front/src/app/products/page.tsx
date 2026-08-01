import type { JSX } from 'react';

import dynamic from 'next/dynamic';

const AdminProductsPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminProductsPage })),
);

const Page = (): JSX.Element => {
  return <AdminProductsPage />;
};

export default Page;
