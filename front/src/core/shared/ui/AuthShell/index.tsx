'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { useEffect } from 'react';

import { authErrorCleared } from '@/core/entities/session';
import { AppPath } from '@/core/shared/router/paths';
import logoImage from '@/core/shared/ui/Header/assets/logo.png';
import logoImage2x from '@/core/shared/ui/Header/assets/logo@2x.png';

import styles from './AuthShell.module.css';

type AuthShellProps = {
  children: React.ReactNode;
  wide?: boolean;
};

export const AuthShell = ({ children, wide = false }: AuthShellProps): JSX.Element => {
  const clearAuthError = useUnit(authErrorCleared);

  useEffect(() => {
    clearAuthError();
  }, [clearAuthError]);

  return (
    <div className={clsx(styles.root)}>
      <div className={clsx(styles.inner, wide && styles.innerWide)}>
        <Link className={clsx(styles.logoLink)} href={AppPath.Home}>
          <img
            alt="Волжский инструмент"
            className={clsx(styles.logoImage)}
            decoding="async"
            height={logoImage.height}
            src={logoImage.src}
            srcSet={`${logoImage.src} 1x, ${logoImage2x.src} 2x`}
            width={logoImage.width}
          />
        </Link>
        <div className={clsx(styles.card)}>{children}</div>
        <Link className={clsx(styles.backLink)} href={AppPath.Home}>
          ← На сайт
        </Link>
      </div>
    </div>
  );
};
