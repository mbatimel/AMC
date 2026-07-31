import dynamic from 'next/dynamic';

const AdminLoginPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminLoginPage })),
);

const Page = (): JSX.Element => {
  return <AdminLoginPage />;
};

export default Page;
