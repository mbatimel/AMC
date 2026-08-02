import dynamic from 'next/dynamic';
import { Suspense } from 'react';

const AdminLoginPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminLoginPage })),
);

const Page = (): JSX.Element => {
  return (
    <Suspense fallback={null}>
      <AdminLoginPage />
    </Suspense>
  );
};

export default Page;
