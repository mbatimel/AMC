'use client';

import clsx from 'clsx';

import styles from './Header.module.css';
import { useHeader } from './model';
import { HeaderDesktop } from './ui/desktop/HeaderDesktop';
import { HeaderMobile } from './ui/mobile/HeaderMobile';

export const Header = (): JSX.Element => {
  const { account, mobileMenu, nav, search } = useHeader();

  return (
    <header className={clsx(styles.root)}>
      <div className={clsx(styles.desktop)}>
        <HeaderDesktop account={account} nav={nav} search={search} />
      </div>

      <div className={clsx(styles.mobile)}>
        <HeaderMobile account={account} mobileMenu={mobileMenu} nav={nav} search={search} />
      </div>
    </header>
  );
};
