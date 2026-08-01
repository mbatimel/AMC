import type { Metadata } from 'next';

import { JSX } from 'react/jsx-runtime';

import '@/core/shared/styles/index.css';
import { ReactScan } from '@/app/ReactScan';
import { ToastViewport } from '@/core/shared/ui/Toast';
import { AdminShell } from '@/views/Admin';

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
