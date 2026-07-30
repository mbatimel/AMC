import dynamic from 'next/dynamic';

const Catalog = dynamic(() =>
  import('@/views/Catalog').then((module) => ({ default: module.Catalog })),
);

const Page = (): JSX.Element => {
  return <Catalog />;
};

export default Page;
