import type { JSX } from 'react';

import dynamic from 'next/dynamic';

const AdminAuditLogPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminAuditLogPage })),
);

const Page = (): JSX.Element => {
  return <AdminAuditLogPage />;
};

export default Page;
