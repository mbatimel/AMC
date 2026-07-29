import dynamic from 'next/dynamic';

const CartPage = dynamic(() =>
  import('@/views/Cart').then((module) => ({ default: module.CartPage })),
);

const Page = (): JSX.Element => {
  return <CartPage />;
};

export default Page;
