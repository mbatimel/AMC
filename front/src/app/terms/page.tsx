import dynamic from 'next/dynamic';

const Terms = dynamic(() =>
  import('@/views/Terms').then((module) => ({ default: module.Terms })),
);

const Page = (): JSX.Element => {
  return <Terms />;
};

export default Page;
