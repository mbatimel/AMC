import type { Icon } from '@/core/shared/icons/types';

import { IconAbout } from '@/core/shared/icons/IconAbout';
import { IconBrands } from '@/core/shared/icons/IconBrands';
import { IconCart } from '@/core/shared/icons/IconCart';
import { IconCatalog } from '@/core/shared/icons/IconCatalog';
import { IconCertificates } from '@/core/shared/icons/IconCertificates';
import { IconContacts } from '@/core/shared/icons/IconContacts';
import { IconHome } from '@/core/shared/icons/IconHome';
import { IconOrders } from '@/core/shared/icons/IconOrders';
import { IconPromotions } from '@/core/shared/icons/IconPromotions';
import { IconSupport } from '@/core/shared/icons/IconSupport';
import { IconTerms } from '@/core/shared/icons/IconTerms';
import { AppPath } from '@/core/shared/router/paths';

export type HeaderMenuItem = {
  href: string;
  icon: Icon;
  label: string;
};

export const HEADER_EMAIL = 'order@amc.ru';

export const HEADER_PHONE_MAIN = '+7 846 265-93-10';

export const HEADER_PHONE_EXTENSIONS = [
  '331-37-10',
  '331-37-11',
  '331-37-12',
  '331-37-13',
  '331-37-15',
];

export const HEADER_SEARCH_PLACEHOLDER = 'Поиск по артикулу, названию, ГОСТ или размеру';

export const HEADER_MOBILE_SEARCH_PLACEHOLDER = 'Поиск товара...';

export const HEADER_ACCOUNT_HREF = AppPath.Cabinet;

export const HEADER_AUTH_HREF = AppPath.Login;

export const HEADER_ACCOUNT_LABEL = 'Кабинет';

export const HEADER_AUTH_LABEL = 'Авторизация';

export const HEADER_CART_HREF = AppPath.Cart;

export const HEADER_ORDERS_HREF = '#';

export const HEADER_NAV_ITEMS: HeaderMenuItem[] = [
  { href: AppPath.Home, icon: IconHome, label: 'Главная' },
  { href: AppPath.Catalog, icon: IconCatalog, label: 'Каталог' },
  { href: '#', icon: IconBrands, label: 'Бренды' },
  { href: '#', icon: IconAbout, label: 'О компании' },
  { href: '#', icon: IconTerms, label: 'Условия работы' },
  { href: '#', icon: IconPromotions, label: 'Акции' },
  { href: '#', icon: IconCertificates, label: 'Сертификаты' },
  { href: '#', icon: IconContacts, label: 'Контакты' },
  { href: '#', icon: IconSupport, label: 'Поддержка' },
];

export const HEADER_CABINET_ITEMS: HeaderMenuItem[] = [
  { href: HEADER_CART_HREF, icon: IconCart, label: 'Корзина' },
  { href: HEADER_ORDERS_HREF, icon: IconOrders, label: 'Мои заказы' },
];
