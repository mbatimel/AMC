import clsx from 'clsx';
import Link from 'next/link';

import type { HeaderMenuItem } from '../../constants';
import type { UseHeaderNavResult } from '../../model/types';

import desktopStyles from '../desktop/HeaderDesktop.module.css';
import mobileStyles from '../mobile/HeaderMobile.module.css';

type HeaderMenuLinkProps = {
  getIsActive: UseHeaderNavResult['getIsActive'];
  item: HeaderMenuItem;
  onNavigate?: () => void;
  variant: HeaderMenuLinkVariant;
};

type HeaderMenuLinkVariant = 'desktop' | 'drawer';

export const HeaderMenuLink = ({
  getIsActive,
  item,
  onNavigate,
  variant,
}: HeaderMenuLinkProps): JSX.Element => {
  const styles = variant === 'desktop' ? desktopStyles : mobileStyles;
  const isActive = getIsActive(item.href);
  const Icon = item.icon;

  if (variant === 'desktop') {
    return (
      <Link
        className={clsx(styles.navLink, isActive && styles.navLinkActive)}
        href={item.href}
        onClick={onNavigate}
      >
        <span>{item.label}</span>
      </Link>
    );
  }

  return (
    <Link
      className={clsx(mobileStyles.drawerLink, isActive && mobileStyles.drawerLinkActive)}
      href={item.href}
      onClick={onNavigate}
    >
      <Icon className={clsx(mobileStyles.drawerLinkIcon)} height={18} width={18} />
      <span>{item.label}</span>
    </Link>
  );
};
