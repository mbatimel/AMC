import dynamic from 'next/dynamic';

const ProductPage = dynamic(() =>
  import('@/views/Product').then((module) => ({ default: module.ProductPage })),
);

const Page = (): JSX.Element => {
  return <ProductPage />;
};

export default Page;
