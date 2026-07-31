'use client';

import { AdminShell } from '@/views/Admin/ui/AdminShell';

type AdminLayoutProps = {
  children: React.ReactNode;
};

const AdminLayout = ({ children }: AdminLayoutProps): JSX.Element => {
  return <AdminShell>{children}</AdminShell>;
};

export default AdminLayout;
