import dynamic from 'next/dynamic';
import { Suspense } from 'react';

const ResetPassword = dynamic(() =>
  import('@/views/ResetPassword').then((module) => ({ default: module.ResetPassword })),
);

const Page = (): JSX.Element => {
  return (
    <Suspense fallback={null}>
      <ResetPassword />
    </Suspense>
  );
};

export default Page;
