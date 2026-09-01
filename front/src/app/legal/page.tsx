import dynamic from 'next/dynamic';

const Legal = dynamic(() => import('@/views/Legal').then((module) => ({ default: module.Legal })));

const Page = (): JSX.Element => {
  return <Legal />;
};

export default Page;
