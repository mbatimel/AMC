import type { Metadata } from 'next';

import { ReactScan } from '@/app/ReactScan';
import '@/core/shared/styles/index.css';

export const metadata: Metadata = {
  title: 'AMC',
};

type RootLayoutProps = {
  children: React.ReactNode;
};

const RootLayout = ({ children }: RootLayoutProps): JSX.Element => {
  return (
    <html lang="ru">
      <ReactScan />
      <body>{children}</body>
    </html>
  );
};

export default RootLayout;
