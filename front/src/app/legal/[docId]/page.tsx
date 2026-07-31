import dynamic from 'next/dynamic';

const Legal = dynamic(() => import('@/views/Legal').then((module) => ({ default: module.Legal })));

type LegalRouteProps = {
  params: Promise<{ docId: string }>;
};

const Page = async ({ params }: LegalRouteProps): Promise<JSX.Element> => {
  const { docId } = await params;

  return <Legal docId={docId} />;
};

export default Page;
