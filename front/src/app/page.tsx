import dynamic from 'next/dynamic';

const Home = dynamic(() => import('@/views/Home').then((module) => ({ default: module.Home })));

const Page = (): JSX.Element => {
  return <Home />;
};

export default Page;
