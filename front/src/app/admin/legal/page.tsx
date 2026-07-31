import dynamic from 'next/dynamic';

const AdminLegalPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminLegalPage })),
);

const Page = (): JSX.Element => {
  return <AdminLegalPage />;
};

export default Page;
