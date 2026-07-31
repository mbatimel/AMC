import dynamic from 'next/dynamic';

const CabinetDocumentsPage = dynamic(() =>
  import('@/views/Cabinet').then((module) => ({ default: module.CabinetDocumentsPage })),
);

const Page = (): JSX.Element => {
  return <CabinetDocumentsPage />;
};

export default Page;
