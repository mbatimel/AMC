'use client';

import { Link as HeroLink } from '@heroui/react';
import clsx from 'clsx';
import { useUnit } from 'effector-react';
import Link from 'next/link';
import { usePathname } from 'next/navigation';

import { useFavorites } from '@/core/entities/favorites';
import { IconFavorite, IconOrders, IconTerms, IconUser } from '@/core/shared/icons';
import { AppPath } from '@/core/shared/router/paths';

import styles from '../Cabinet.module.css';
import { $ordersTotal } from '../model/orders';
import { $conditions, $profile } from '../model/profile';

const NAV_ITEMS = [
  {
    badgeKey: 'orders' as const,
    href: AppPath.CabinetOrders,
    icon: IconOrders,
    label: 'Мои заказы',
  },
  {
    badgeKey: 'documents' as const,
    href: '#documents',
    icon: IconTerms,
    label: 'Документы',
  },
  {
    badgeKey: 'favorites' as const,
    href: '#favorites',
    icon: IconFavorite,
    label: 'Избранное',
  },
  {
    badgeKey: null,
    href: AppPath.CabinetProfile,
    icon: IconUser,
    label: 'Профиль и условия',
  },
] as const;

export const CabinetSidebar = (): JSX.Element => {
  const pathname = usePathname();
  const [profile, conditions, ordersTotal] = useUnit([$profile, $conditions, $ordersTotal]);
  const { favoriteIds } = useFavorites();

  const companyName =
    profile?.active_client?.company_name ||
    [profile?.last_name, profile?.first_name].filter(Boolean).join(' ') ||
    'Личный кабинет';
  const groupLabel = conditions?.price_group
    ? `Группа: ${conditions.price_group}`
    : profile?.active_client?.company_type
      ? `Группа: ${profile.active_client.company_type}`
      : null;

  const badges: Record<'documents' | 'favorites' | 'orders', number> = {
    documents: 0,
    favorites: favoriteIds.length,
    orders: ordersTotal,
  };

  return (
    <aside className={clsx(styles.sidebar)}>
      <div className={clsx(styles.sidebarHeader)}>
        <p className={clsx(styles.sidebarCompany)}>{companyName}</p>
        {groupLabel ? <p className={clsx(styles.sidebarGroup)}>{groupLabel}</p> : null}
      </div>

      <nav aria-label="Разделы кабинета" className={clsx(styles.sidebarNav)}>
        {NAV_ITEMS.map((item) => {
          const href = item.href as string;
          const isRoute = href.startsWith('/');
          const isActive = isRoute && (pathname === href || pathname.startsWith(`${href}/`));
          const Icon = item.icon;
          const badgeValue = item.badgeKey ? badges[item.badgeKey] : 0;
          const iconColor = isActive ? 'var(--header-brand)' : '#8B93A7';

          const content = (
            <>
              <Icon
                className={clsx(styles.sidebarIcon)}
                currentColor={iconColor}
                height={18}
                width={18}
              />
              <span className={clsx(styles.sidebarLinkLabel)}>{item.label}</span>
              {badgeValue > 0 ? (
                <span className={clsx(styles.sidebarBadge)}>{badgeValue}</span>
              ) : null}
            </>
          );

          if (!isRoute) {
            return (
              <span className={clsx(styles.sidebarLink, styles.sidebarLinkMuted)} key={item.label}>
                {content}
              </span>
            );
          }

          return (
            <Link
              className={clsx(styles.sidebarLink, isActive && styles.sidebarLinkActive)}
              href={item.href}
              key={item.href}
            >
              {content}
            </Link>
          );
        })}
      </nav>

      <div className={clsx(styles.sidebarFooter)}>
        <HeroLink className={clsx(styles.sidebarHome)} href={AppPath.Home}>
          <span aria-hidden className={clsx(styles.sidebarHomeArrow)}>
            ←
          </span>
          На главную
        </HeroLink>
      </div>
    </aside>
  );
};
