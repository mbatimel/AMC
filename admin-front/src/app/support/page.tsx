import type { JSX } from 'react';

import dynamic from 'next/dynamic';

const AdminSupportPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminSupportPage })),
);

const Page = (): JSX.Element => {
  return <AdminSupportPage />;
};

export default Page;
