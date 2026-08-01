import type { JSX } from 'react';

import dynamic from 'next/dynamic';

import type { ContentPageKey } from '@/core/shared/api/content';

const AdminContentPage = dynamic(() =>
  import('@/views/Admin').then((module) => ({ default: module.AdminContentPage })),
);

type AdminContentRouteProps = {
  params: Promise<{ pageKey: string }>;
};

const PAGE_KEYS: ContentPageKey[] = ['about', 'certificates', 'contacts', 'home', 'promo', 'terms'];

const Page = async ({ params }: AdminContentRouteProps): Promise<JSX.Element> => {
  const { pageKey } = await params;
  const key = PAGE_KEYS.includes(pageKey as ContentPageKey)
    ? (pageKey as ContentPageKey)
    : 'home';

  return <AdminContentPage pageKey={key} />;
};

export default Page;
