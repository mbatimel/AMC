import dynamic from 'next/dynamic';

const AdminDashboardPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminDashboardPage })),
);

const Page = (): JSX.Element => {
  return <AdminDashboardPage />;
};

export default Page;
