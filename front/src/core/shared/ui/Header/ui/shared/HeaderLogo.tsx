import clsx from 'clsx';
import Link from 'next/link';

import { AppPath } from '@/core/shared/router/paths';

import logoImage from '../../assets/logo.png';
import logoImage2x from '../../assets/logo@2x.png';
import styles from '../../Header.module.css';

type HeaderLogoProps = {
  className?: string;
};

export const HeaderLogo = ({ className }: HeaderLogoProps): JSX.Element => {
  return (
    <Link className={clsx(styles.logoLink, className)} href={AppPath.Home}>
      <img
        alt="Волжский инструмент"
        className={clsx(styles.logoImage)}
        decoding="async"
        fetchPriority="high"
        height={logoImage.height}
        loading="eager"
        src={logoImage.src}
        srcSet={`${logoImage.src} 1x, ${logoImage2x.src} 2x`}
        width={logoImage.width}
      />
    </Link>
  );
};
