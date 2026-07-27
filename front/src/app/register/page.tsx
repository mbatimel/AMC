import dynamic from 'next/dynamic';

const Register = dynamic(() =>
  import('@/views/Register').then((module) => ({ default: module.Register })),
);

const Page = (): JSX.Element => {
  return <Register />;
};

export default Page;
