import dynamic from 'next/dynamic';

const AdminPromotionsPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminPromotionsPage })),
);

const Page = (): JSX.Element => {
  return <AdminPromotionsPage />;
};

export default Page;
