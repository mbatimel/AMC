import type { Metadata } from 'next';

import { EffectorProvider } from '@/app/providers/EffectorProvider';
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
      <body>
        <EffectorProvider>{children}</EffectorProvider>
      </body>
    </html>
  );
};

export default RootLayout;
