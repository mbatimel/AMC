import type { Metadata } from 'next';

import { ReactScan } from '@/app/ReactScan';
import '@/core/shared/styles/index.css';
import { AdminShell } from '@/views/Admin';
import { ToastViewport } from '@/core/shared/ui/Toast';

export const metadata: Metadata = {
  title: 'AMC Admin',
};

type RootLayoutProps = {
  children: React.ReactNode;
};

const RootLayout = ({ children }: RootLayoutProps): JSX.Element => {
  return (
    <html lang="ru">
      <ReactScan />
      <body>
        <AdminShell>{children}</AdminShell>
        <ToastViewport />
      </body>
    </html>
  );
};

export default RootLayout;
