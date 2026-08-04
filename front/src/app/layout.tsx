import type { Metadata } from 'next';

import { ReactScan } from '@/app/ReactScan';
import '@/core/shared/styles/index.css';
import { ToastViewport } from '@/core/shared/ui/Toast';

export const metadata: Metadata = {
  title: 'VOINT',
};

type RootLayoutProps = {
  children: React.ReactNode;
};

const RootLayout = ({ children }: RootLayoutProps): JSX.Element => {
  return (
    <html lang="ru">
      <ReactScan />
      <body>
        {children}
        <ToastViewport />
      </body>
    </html>
  );
};

export default RootLayout;
