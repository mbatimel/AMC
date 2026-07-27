import dynamic from 'next/dynamic';

const ForgotPassword = dynamic(() =>
  import('@/views/ForgotPassword').then((module) => ({ default: module.ForgotPassword })),
);

const Page = (): JSX.Element => {
  return <ForgotPassword />;
};

export default Page;
