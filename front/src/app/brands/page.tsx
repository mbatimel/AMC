import dynamic from 'next/dynamic';

const Brands = dynamic(() =>
  import('@/views/Brands').then((module) => ({ default: module.Brands })),
);

const Page = (): JSX.Element => {
  return <Brands />;
};

export default Page;
