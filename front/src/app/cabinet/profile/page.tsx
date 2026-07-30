import dynamic from 'next/dynamic';

const CabinetProfilePage = dynamic(() =>
  import('@/views/Cabinet').then((module) => ({ default: module.CabinetProfilePage })),
);

const Page = (): JSX.Element => {
  return <CabinetProfilePage />;
};

export default Page;
