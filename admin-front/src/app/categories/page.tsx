import type { JSX } from 'react';

import dynamic from 'next/dynamic';

const AdminCategoriesPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminCategoriesPage })),
);

const Page = (): JSX.Element => {
  return <AdminCategoriesPage />;
};

export default Page;
