import dynamic from 'next/dynamic';

const AdminUsersPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminUsersPage })),
);

const Page = (): JSX.Element => {
  return <AdminUsersPage />;
};

export default Page;
