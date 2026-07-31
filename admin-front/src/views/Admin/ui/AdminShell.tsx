'use client';

import clsx from 'clsx';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useEffect } from 'react';

import { useAdminSession } from '@/core/entities/adminSession';
import { AppPath } from '@/core/shared/router/paths';

import styles from '../Admin.module.css';
import { ADMIN_NAV } from '../lib/nav';

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
  const isLoginRoute = pathname === AppPath.Login;

  useEffect(() => {
    if (isHydrated && !isAuthenticated && !isLoginRoute) {
      router.replace(AppPath.Login);
    }
  }, [isAuthenticated, isHydrated, isLoginRoute, router]);

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
          <Link className={clsx(styles.topbarLink)} href={AppPath.Home}>
            На сайт
          </Link>
          <button
            className={clsx(styles.topbarButton)}
            onClick={() => {
              logout();
              router.replace(AppPath.Login);
            }}
            type="button"
          >
            Выйти
          </button>
        </div>
      </header>

      <div className={clsx(styles.body)}>
        <aside className={clsx(styles.sidebar)}>
          {ADMIN_NAV.map((group) => (
            <nav className={clsx(styles.navGroup)} key={group.title ?? 'root'}>
              {group.title ? <p className={clsx(styles.navTitle)}>{group.title}</p> : null}
              {group.items.map((item) => {
                const isActive =
                  item.href === AppPath.Home
                    ? pathname === AppPath.Home
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
