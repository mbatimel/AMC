import dynamic from 'next/dynamic';
import { Suspense } from 'react';

const Login = dynamic(() => import('@/views/Login').then((module) => ({ default: module.Login })));

const Page = (): JSX.Element => {
  return (
    <Suspense fallback={null}>
      <Login />
    </Suspense>
  );
};

export default Page;
