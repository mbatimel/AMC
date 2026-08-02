'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { adminAuthErrorCleared } from '@/core/entities/adminSession';
import logoImage from '@/core/shared/assets/logo.png';
import logoImage2x from '@/core/shared/assets/logo@2x.png';
import { PUBLIC_SITE_ORIGIN } from '@/core/shared/router/paths';

import styles from './AuthShell.module.css';

type AuthShellProps = {
  children: React.ReactNode;
  wide?: boolean;
};

export const AuthShell = ({ children, wide = false }: AuthShellProps): JSX.Element => {
  const clearAuthError = useUnit(adminAuthErrorCleared);

  useEffect(() => {
    clearAuthError();
  }, [clearAuthError]);

  return (
    <div className={clsx(styles.root)}>
      <div className={clsx(styles.inner, wide && styles.innerWide)}>
        <a className={clsx(styles.logoLink)} href={PUBLIC_SITE_ORIGIN}>
          <img
            alt="Волжский инструмент"
            className={clsx(styles.logoImage)}
            decoding="async"
            height={logoImage.height}
            src={logoImage.src}
            srcSet={`${logoImage.src} 1x, ${logoImage2x.src} 2x`}
            width={logoImage.width}
          />
        </a>
        <div className={clsx(styles.card)}>{children}</div>
        <a className={clsx(styles.backLink)} href={PUBLIC_SITE_ORIGIN}>
          ← На сайт
        </a>
      </div>
    </div>
  );
};
