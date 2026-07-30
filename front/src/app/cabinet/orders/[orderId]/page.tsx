import dynamic from 'next/dynamic';

const CabinetOrderPage = dynamic(() =>
  import('@/views/Cabinet').then((module) => ({ default: module.CabinetOrderPage })),
);

type PageProps = {
  params: Promise<{ orderId: string }>;
};

const Page = async ({ params }: PageProps): Promise<JSX.Element> => {
  const { orderId } = await params;

  return <CabinetOrderPage orderId={orderId} />;
};

export default Page;
