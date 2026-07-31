import dynamic from 'next/dynamic';

const Promo = dynamic(() =>
  import('@/views/Promo').then((module) => ({ default: module.Promo })),
);

const Page = (): JSX.Element => {
  return <Promo />;
};

export default Page;
