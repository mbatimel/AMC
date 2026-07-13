import clsx from 'clsx';

import styles from './Footer.module.css';
import { FooterDesktop } from './ui/desktop/FooterDesktop';
import { FooterMobile } from './ui/mobile/FooterMobile';

export const Footer = (): JSX.Element => {
  return (
    <footer className={clsx(styles.root)}>
      <div className={clsx(styles.desktop)}>
        <FooterDesktop />
      </div>

      <div className={clsx(styles.mobile)}>
        <FooterMobile />
      </div>
    </footer>
  );
};
