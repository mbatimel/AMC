import clsx from 'clsx';

import {
  IconCart,
  IconChevronDown,
  IconEmail,
  IconLocation,
  IconOrders,
  IconPhone,
  IconUser,
} from '@/core/shared/icons';

import type { HeaderViewProps } from '../../model/types';

import {
  HEADER_ACCOUNT_HREF,
  HEADER_CART_HREF,
  HEADER_CITY,
  HEADER_EMAIL,
  HEADER_ORDERS_HREF,
  HEADER_PHONE_EXTENSIONS,
  HEADER_PHONE_MAIN,
} from '../../constants';
import rootStyles from '../../Header.module.css';
import { HeaderLogo } from '../shared/HeaderLogo';
import { HeaderMenuLink } from '../shared/HeaderMenuLink';
import { HeaderSearchForm } from '../shared/HeaderSearchForm';
import styles from './HeaderDesktop.module.css';

export const HeaderDesktopTopBar = (): JSX.Element => {
  return (
    <div className={clsx(styles.topBar)}>
      <div className={clsx(rootStyles.container, styles.topBarInner)}>
        <button className={clsx(styles.cityButton)} type="button">
          <IconLocation className={clsx(styles.cityIcon)} height={14} width={14} />
          <span>{HEADER_CITY}</span>
          <IconChevronDown className={clsx(styles.cityChevron)} height={12} width={12} />
        </button>

        <div className={clsx(styles.contacts)}>
          <a className={clsx(styles.contactLink)} href={`mailto:${HEADER_EMAIL}`}>
            <IconEmail className={clsx(styles.contactIcon)} height={14} width={14} />
            <span>{HEADER_EMAIL}</span>
          </a>

          <div className={clsx(styles.phones)}>
            <IconPhone className={clsx(styles.contactIcon)} height={14} width={14} />
            <a
              className={clsx(styles.contactLink)}
              href={`tel:${HEADER_PHONE_MAIN.replace(/\s/g, '')}`}
            >
              <span>{HEADER_PHONE_MAIN}</span>
            </a>
            {HEADER_PHONE_EXTENSIONS.map((phone) => (
              <span className={clsx(styles.phoneExtension)} key={phone}>
                <span aria-hidden="true" className={clsx(styles.phoneDivider)} />
                <a className={clsx(styles.contactLink)} href={`tel:${phone.replace(/-/g, '')}`}>
                  <span>{phone}</span>
                </a>
              </span>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export const HeaderDesktopMain = ({ search }: Pick<HeaderViewProps, 'search'>): JSX.Element => {
  return (
    <div className={clsx(styles.mainBar)}>
      <div className={clsx(rootStyles.container, styles.mainBarInner)}>
        <HeaderLogo />

        <HeaderSearchForm search={search} variant="desktop" />

        <div className={clsx(styles.actions)}>
          <a className={clsx(styles.cartButton)} href={HEADER_CART_HREF}>
            <IconCart height={16} width={16} />
            <span>Корзина</span>
          </a>

          <div className={clsx(styles.accountBlock)}>
            <a className={clsx(styles.accountLink)} href={HEADER_ACCOUNT_HREF}>
              <IconUser height={16} width={16} />
              <span>Кабинет</span>
            </a>
            <a className={clsx(styles.accountLink)} href={HEADER_ORDERS_HREF}>
              <IconOrders height={16} width={16} />
              <span>Мои заказы</span>
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};

export const HeaderDesktopNav = ({ nav }: Pick<HeaderViewProps, 'nav'>): JSX.Element => {
  return (
    <nav aria-label="Основное меню" className={clsx(styles.navBar)}>
      <div className={clsx(rootStyles.container, styles.navBarInner)}>
        <ul className={clsx(styles.navList)}>
          {nav.navItems.map((item) => (
            <li className={clsx(styles.navItem)} key={item.label}>
              <HeaderMenuLink getIsActive={nav.getIsActive} item={item} variant="desktop" />
            </li>
          ))}
        </ul>
      </div>
    </nav>
  );
};

export const HeaderDesktop = ({ nav, search }: HeaderViewProps): JSX.Element => {
  return (
    <>
      <HeaderDesktopTopBar />
      <HeaderDesktopMain search={search} />
      <HeaderDesktopNav nav={nav} />
    </>
  );
};
