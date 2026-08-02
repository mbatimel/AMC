import dynamic from 'next/dynamic';

const AdminUserDetailPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminUserDetailPage })),
);

type AdminUserDetailRouteProps = {
  params: Promise<{ userId: string }>;
};

const Page = async ({ params }: AdminUserDetailRouteProps): Promise<JSX.Element> => {
  const { userId } = await params;

  return <AdminUserDetailPage userId={userId} />;
};

export default Page;
