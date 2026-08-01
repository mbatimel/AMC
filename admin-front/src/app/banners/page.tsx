import type { JSX } from 'react';

import dynamic from 'next/dynamic';

const AdminBannersPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminBannersPage })),
);

const Page = (): JSX.Element => {
  return <AdminBannersPage />;
};

export default Page;
