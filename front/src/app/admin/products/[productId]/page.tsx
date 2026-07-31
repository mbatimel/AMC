import dynamic from 'next/dynamic';

const AdminProductPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminProductPage })),
);

type AdminProductRouteProps = {
  params: Promise<{ productId: string }>;
};

const Page = async ({ params }: AdminProductRouteProps): Promise<JSX.Element> => {
  const { productId } = await params;

  return <AdminProductPage productId={productId} />;
};

export default Page;
