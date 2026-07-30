import dynamic from 'next/dynamic';

const CheckoutPage = dynamic(() =>
  import('@/views/Checkout').then((module) => ({ default: module.CheckoutPage })),
);

const Page = (): JSX.Element => {
  return <CheckoutPage />;
};

export default Page;
