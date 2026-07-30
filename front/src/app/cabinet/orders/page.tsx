import dynamic from 'next/dynamic';

const CabinetOrdersPage = dynamic(() =>
  import('@/views/Cabinet').then((module) => ({ default: module.CabinetOrdersPage })),
);

const Page = (): JSX.Element => {
  return <CabinetOrdersPage />;
};

export default Page;
