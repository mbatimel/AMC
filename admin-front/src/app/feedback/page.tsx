import type { JSX } from 'react';

import dynamic from 'next/dynamic';

const AdminFeedbackPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminFeedbackPage })),
);

const Page = (): JSX.Element => {
  return <AdminFeedbackPage />;
};

export default Page;
