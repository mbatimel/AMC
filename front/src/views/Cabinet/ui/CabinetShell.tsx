'use client';

import { Breadcrumbs } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { AppPath } from '@/core/shared/router/paths';
import { Page } from '@/core/shared/ui/Page';
import { ToastViewport } from '@/core/shared/ui/Toast';

import styles from '../Cabinet.module.css';
import { cabinetOrdersOpened } from '../model/orders';
import { cabinetProfileOpened } from '../model/profile';
import { CabinetSidebar } from './CabinetSidebar';

type CabinetShellProps = {
  children: React.ReactNode;
};

export const CabinetShell = ({ children }: CabinetShellProps): JSX.Element => {
  const [openOrders, openProfile] = useUnit([cabinetOrdersOpened, cabinetProfileOpened]);

  useEffect(() => {
    openProfile();
    openOrders();
  }, [openOrders, openProfile]);

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <div className={clsx(styles.container)}>
          <Breadcrumbs className={clsx(styles.breadcrumbs)}>
            <Breadcrumbs.Item href={AppPath.Home}>Главная</Breadcrumbs.Item>
            <Breadcrumbs.Item>Личный кабинет</Breadcrumbs.Item>
          </Breadcrumbs>

          <div className={clsx(styles.layout)}>
            <CabinetSidebar />
            <div className={clsx(styles.content)}>{children}</div>
          </div>
        </div>
        <ToastViewport />
      </div>
    </Page>
  );
};
