import type { Metadata } from 'next';

import { ReactScan } from '@/app/ReactScan';
import '@/core/shared/styles/index.css';
import { ToastViewport } from '@/core/shared/ui/Toast';
import { AdminShell } from '@/views/Admin';

export const metadata: Metadata = {
  title: 'VOINT Admin',
};

type RootLayoutProps = {
  children: React.ReactNode;
};

const RootLayout = ({ children }: RootLayoutProps): JSX.Element => {
  return (
    <html lang="ru">
      {process.env.NODE_ENV === 'development' ? <ReactScan /> : null}
      <body>
        <AdminShell>{children}</AdminShell>
        <ToastViewport />
      </body>
    </html>
  );
};

export default RootLayout;
