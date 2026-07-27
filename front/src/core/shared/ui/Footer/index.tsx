'use client';

import clsx from 'clsx';

import { useSession } from '@/core/entities/session';
import { AppPath } from '@/core/shared/router/paths';

import type { FooterLinkItem } from './constants';

import { FOOTER_CABINET_LINKS } from './constants';
import styles from './Footer.module.css';
import { FooterDesktop } from './ui/desktop/FooterDesktop';
import { FooterMobile } from './ui/mobile/FooterMobile';

const getCabinetLinks = (isAuthenticated: boolean): FooterLinkItem[] => {
  if (isAuthenticated) {
    return FOOTER_CABINET_LINKS;
  }

  return [
    { href: AppPath.Login, label: 'Авторизация' },
    ...FOOTER_CABINET_LINKS.slice(1),
  ];
};

export const Footer = (): JSX.Element => {
  const { isAuthenticated } = useSession();
  const cabinetLinks = getCabinetLinks(isAuthenticated);
  const cabinetTitle = isAuthenticated ? 'Кабинет' : 'Аккаунт';

  return (
    <footer className={clsx(styles.root)}>
      <div className={clsx(styles.desktop)}>
        <FooterDesktop cabinetLinks={cabinetLinks} cabinetTitle={cabinetTitle} />
      </div>

      <div className={clsx(styles.mobile)}>
        <FooterMobile cabinetLinks={cabinetLinks} cabinetTitle={cabinetTitle} />
      </div>
    </footer>
  );
};
