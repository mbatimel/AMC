import dynamic from 'next/dynamic';

const Contacts = dynamic(() =>
  import('@/views/Contacts').then((module) => ({ default: module.Contacts })),
);

const Page = (): JSX.Element => {
  return <Contacts />;
};

export default Page;
