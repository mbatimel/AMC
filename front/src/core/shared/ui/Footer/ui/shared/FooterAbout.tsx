import clsx from 'clsx';
import Link from 'next/link';

import { AppPath } from '@/core/shared/router/paths';
import logoImage from '@/core/shared/ui/Header/assets/logo.png';
import logoImage2x from '@/core/shared/ui/Header/assets/logo@2x.png';

import { FOOTER_DESCRIPTION } from '../../constants';
import styles from '../../Footer.module.css';

export const FooterLogo = (): JSX.Element => {
  return (
    <Link className={clsx(styles.logoLink)} href={AppPath.Home}>
      <img
        alt="Волжский инструмент"
        className={clsx(styles.logoImage)}
        decoding="async"
        height={logoImage.height}
        loading="lazy"
        src={logoImage.src}
        srcSet={`${logoImage.src} 1x, ${logoImage2x.src} 2x`}
        width={logoImage.width}
      />
    </Link>
  );
};

export const FooterAbout = (): JSX.Element => {
  return (
    <div>
      <FooterLogo />
      <p className={clsx(styles.description)}>
        <span>{FOOTER_DESCRIPTION}</span>
      </p>
    </div>
  );
};
