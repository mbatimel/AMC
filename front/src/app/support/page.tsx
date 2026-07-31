import dynamic from 'next/dynamic';

const Support = dynamic(() =>
  import('@/views/Support').then((module) => ({ default: module.Support })),
);

const Page = (): JSX.Element => {
  return <Support />;
};

export default Page;
