import type { JSX } from 'react';

import dynamic from 'next/dynamic';

const AdminSignupRequestsPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminSignupRequestsPage })),
);

const Page = (): JSX.Element => {
  return <AdminSignupRequestsPage />;
};

export default Page;
