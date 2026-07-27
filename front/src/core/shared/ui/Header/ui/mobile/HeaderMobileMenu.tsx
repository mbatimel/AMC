import clsx from 'clsx';

import { IconClose, IconLogin, IconUser } from '@/core/shared/icons';

import type { HeaderMenuItem } from '../../constants';
import type { HeaderMobileMenuProps } from '../../model/types';

import { HeaderLogo } from '../shared/HeaderLogo';
import { HeaderMenuLink } from '../shared/HeaderMenuLink';
import { HeaderSearchForm } from '../shared/HeaderSearchForm';
import styles from './HeaderMobile.module.css';

export const HeaderMobileMenu = ({
  account,
  closeMenu,
  drawerId,
  drawerRef,
  isOpen,
  nav,
  search,
}: HeaderMobileMenuProps): JSX.Element | null => {
  if (!isOpen) {
    return null;
  }

  const onNavigate = (): void => {
    closeMenu();
  };

  const accountItem: HeaderMenuItem = {
    href: account.href,
    icon: account.isAuthenticated ? IconUser : IconLogin,
    label: account.label,
  };

  return (
    <>
      <button
        aria-label="Закрыть меню"
        className={clsx(styles.overlay)}
        onClick={closeMenu}
        tabIndex={-1}
        type="button"
      />

      <aside
        aria-label="Мобильное меню"
        aria-modal="true"
        className={clsx(styles.drawer)}
        id={drawerId}
        ref={drawerRef}
        role="dialog"
      >
        <div className={clsx(styles.drawerHeader)}>
          <HeaderLogo />
          <button
            aria-label="Закрыть меню"
            className={clsx(styles.closeButton)}
            onClick={closeMenu}
            type="button"
          >
            <IconClose height={20} width={20} />
          </button>
        </div>

        <div className={clsx(styles.drawerSearch)}>
          <HeaderSearchForm search={search} variant="drawer" />
        </div>

        <section className={clsx(styles.drawerSection)}>
          <div className={clsx(styles.drawerSectionTitle)}>
            <span>Навигация</span>
          </div>
          <div className={clsx(styles.drawerLinks)}>
            {nav.navItems.map((item) => (
              <HeaderMenuLink
                getIsActive={nav.getIsActive}
                item={item}
                key={item.label}
                onNavigate={onNavigate}
                variant="drawer"
              />
            ))}
          </div>
        </section>

        <section className={clsx(styles.drawerSection)}>
          <div className={clsx(styles.drawerSectionTitle)}>
            <span>{account.isAuthenticated ? 'Кабинет' : 'Аккаунт'}</span>
          </div>
          <div className={clsx(styles.drawerLinks)}>
            <HeaderMenuLink
              getIsActive={nav.getIsActive}
              item={accountItem}
              onNavigate={onNavigate}
              variant="drawer"
            />
            {account.isAuthenticated
              ? nav.cabinetItems.map((item) => (
                  <HeaderMenuLink
                    getIsActive={nav.getIsActive}
                    item={item}
                    key={item.label}
                    onNavigate={onNavigate}
                    variant="drawer"
                  />
                ))
              : null}
          </div>
        </section>
      </aside>
    </>
  );
};
