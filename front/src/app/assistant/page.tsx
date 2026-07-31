import dynamic from 'next/dynamic';

const Assistant = dynamic(() =>
  import('@/views/Assistant').then((module) => ({ default: module.Assistant })),
);

const Page = (): JSX.Element => {
  return <Assistant />;
};

export default Page;
