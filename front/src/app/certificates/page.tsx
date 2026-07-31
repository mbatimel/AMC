import dynamic from 'next/dynamic';

const Certificates = dynamic(() =>
  import('@/views/Certificates').then((module) => ({ default: module.Certificates })),
);

const Page = (): JSX.Element => {
  return <Certificates />;
};

export default Page;
