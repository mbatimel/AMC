import dynamic from 'next/dynamic';

const Cabinet = dynamic(() =>
  import('@/views/Cabinet').then((module) => ({ default: module.Cabinet })),
);

const Page = (): JSX.Element => {
  return <Cabinet />;
};

export default Page;
