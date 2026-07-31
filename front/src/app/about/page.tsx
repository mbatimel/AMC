import dynamic from 'next/dynamic';

const About = dynamic(() =>
  import('@/views/About').then((module) => ({ default: module.About })),
);

const Page = (): JSX.Element => {
  return <About />;
};

export default Page;
