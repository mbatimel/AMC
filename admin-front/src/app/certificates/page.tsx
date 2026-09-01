import dynamic from 'next/dynamic';

const AdminCertificatesPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminCertificatesPage })),
);

const Page = (): JSX.Element => {
  return <AdminCertificatesPage />;
};

export default Page;
