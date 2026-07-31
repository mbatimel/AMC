import dynamic from 'next/dynamic';

const CabinetFavoritesPage = dynamic(() =>
  import('@/views/Cabinet').then((module) => ({ default: module.CabinetFavoritesPage })),
);

const Page = (): JSX.Element => {
  return <CabinetFavoritesPage />;
};

export default Page;
