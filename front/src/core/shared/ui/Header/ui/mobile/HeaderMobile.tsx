import clsx from 'clsx';
import Link from 'next/link';

import { IconCart, IconLogin, IconMenu, IconOrders, IconUser } from '@/core/shared/icons';

import type { HeaderMobileViewProps } from '../../model/types';

import { HEADER_ORDERS_HREF } from '../../constants';
import rootStyles from '../../Header.module.css';
import { HeaderLogo } from '../shared/HeaderLogo';
import { HeaderSearchForm } from '../shared/HeaderSearchForm';
import styles from './HeaderMobile.module.css';
import { HeaderMobileMenu } from './HeaderMobileMenu';

export const HeaderMobile = ({
  account,
  cart,
  mobileMenu,
  nav,
  search,
}: HeaderMobileViewProps): JSX.Element => {
  const { closeMenu, drawerId, drawerRef, isOpen, menuTriggerRef, openMenu } = mobileMenu;
  const AccountIcon = account.isAuthenticated ? IconUser : IconLogin;

  return (
    <>
      <div className={clsx(styles.toolbar)}>
        <div className={clsx(rootStyles.container, styles.toolbarInner)}>
          <button
            aria-controls={drawerId}
            aria-expanded={isOpen}
            aria-haspopup="dialog"
            aria-label="Открыть меню"
            className={clsx(styles.menuButton)}
            onClick={openMenu}
            ref={menuTriggerRef}
            type="button"
          >
            <IconMenu height={18} width={18} />
          </button>

          <HeaderLogo />

          <div className={clsx(styles.toolbarActions)}>
            <Link
              aria-label={cart.count > 0 ? `Корзина, ${cart.count}` : 'Корзина'}
              className={clsx(styles.iconButton, styles.iconButtonCart)}
              href={cart.href}
            >
              <IconCart height={18} width={18} />
              {cart.count > 0 ? <span className={clsx(styles.cartBadge)}>{cart.count}</span> : null}
            </Link>
            <a aria-label={account.label} className={clsx(styles.iconButton)} href={account.href}>
              <AccountIcon height={18} width={18} />
            </a>
            <a
              aria-label="Мои заказы"
              className={clsx(styles.iconButton)}
              href={HEADER_ORDERS_HREF}
            >
              <IconOrders height={18} width={18} />
            </a>
          </div>
        </div>
      </div>

      <div className={clsx(rootStyles.container, styles.searchRow)}>
        <HeaderSearchForm search={search} variant="mobile" />
      </div>

      <HeaderMobileMenu
        account={account}
        closeMenu={closeMenu}
        drawerId={drawerId}
        drawerRef={drawerRef}
        isOpen={isOpen}
        nav={nav}
        search={search}
      />
    </>
  );
};
