'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect } from 'react';

import { useAdminSession } from '@/core/entities/adminSession';
import { AppPath, PUBLIC_SITE_ORIGIN } from '@/core/shared/router/paths';

import { ADMIN_NAV } from '../lib/nav';
import styles from './AdminShell.module.css';

type AdminShellProps = {
  children: React.ReactNode;
};

/**
 * Каркас админ-панели. Страница входа рендерится без сайдбара,
 * остальные разделы требуют активной admin-сессии.
 */
export const AdminShell = ({ children }: AdminShellProps): JSX.Element => {
  const pathname = usePathname();
  const router = useRouter();
  const { isAuthenticated, isHydrated, logout } = useAdminSession();
  const isLoginRoute = pathname === (AppPath.Login as string);

  useEffect(() => {
    if (isHydrated && !isAuthenticated && !isLoginRoute) {
      const next = pathname && pathname !== (AppPath.Login as string) ? pathname : AppPath.Home;
      router.replace(`${AppPath.Login}?next=${encodeURIComponent(next)}`);
    }
  }, [isAuthenticated, isHydrated, isLoginRoute, pathname, router]);

  if (isLoginRoute) {
    return <>{children}</>;
  }

  if (!isHydrated || !isAuthenticated) {
    return (
      <div className={clsx(styles.gate)}>
        <p>Проверяем доступ…</p>
      </div>
    );
  }

  return (
    <div className={clsx(styles.shell)}>
      <header className={clsx(styles.topbar)}>
        <Link className={clsx(styles.brand)} href={AppPath.Home}>
          ВИ-Портал · Админ-панель
        </Link>
        <div className={clsx(styles.topbarActions)}>
          <a className={clsx(styles.topbarLink)} href={PUBLIC_SITE_ORIGIN}>
            На сайт
          </a>
          <Button
            className={clsx(styles.topbarButton)}
            onPress={() => {
              void logout().finally(() => {
                window.location.assign(AppPath.Login);
              });
            }}
            type="button"
            variant="secondary"
          >
            Выйти
          </Button>
        </div>
      </header>

      <div className={clsx(styles.body)}>
        <aside className={clsx(styles.sidebar)}>
          {ADMIN_NAV.map((group) => (
            <nav className={clsx(styles.navGroup)} key={group.title ?? 'root'}>
              {group.title ? <p className={clsx(styles.navTitle)}>{group.title}</p> : null}
              {group.items.map((item) => {
                const isActive =
                  item.href === (AppPath.Home as string)
                    ? pathname === (AppPath.Home as string)
                    : pathname.startsWith(item.href);

                return (
                  <Link
                    className={clsx(styles.navLink, isActive && styles.navLinkActive)}
                    href={item.href}
                    key={item.href}
                  >
                    {item.label}
                  </Link>
                );
              })}
            </nav>
          ))}
        </aside>

        <main className={clsx(styles.content)}>{children}</main>
      </div>
    </div>
  );
};
