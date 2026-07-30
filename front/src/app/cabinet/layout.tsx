'use client';

import { CabinetShell } from '@/views/Cabinet/ui/CabinetShell';

type CabinetLayoutProps = {
  children: React.ReactNode;
};

const CabinetLayout = ({ children }: CabinetLayoutProps): JSX.Element => {
  return <CabinetShell>{children}</CabinetShell>;
};

export default CabinetLayout;
