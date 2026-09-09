import { AppPath } from '@/core/shared/router/paths';

export type FooterLinkItem = {
  href: string;
  label: string;
};

export const FOOTER_BRANDS_TITLE = 'Бренды на портале';

export const FOOTER_DESCRIPTION =
  'Производитель и оптовый поставщик профессионального инструмента.';

export const FOOTER_CATALOG_LINKS: FooterLinkItem[] = [
  { href: AppPath.Catalog, label: 'Металлорежущий' },
  { href: AppPath.Catalog, label: 'Метчики' },
  { href: AppPath.Catalog, label: 'Сверла' },
  { href: AppPath.Catalog, label: 'Плашки' },
  { href: AppPath.Catalog, label: 'Слесарный' },
  { href: AppPath.Catalog, label: 'Садовый' },
];

export const FOOTER_CABINET_LINKS: FooterLinkItem[] = [
  { href: AppPath.Cabinet, label: 'Личный кабинет' },
  { href: AppPath.CabinetOrders, label: 'Мои заказы' },
  { href: AppPath.CabinetDocuments, label: 'Документы' },
  { href: AppPath.CabinetFavorites, label: 'Избранное' },
  { href: AppPath.CabinetProfile, label: 'Профиль' },
];

export const FOOTER_INFO_LINKS: FooterLinkItem[] = [
  { href: AppPath.About, label: 'О компании' },
  { href: AppPath.Terms, label: 'Условия работы' },
  { href: AppPath.Certificates, label: 'Сертификаты' },
  { href: AppPath.Contacts, label: 'Контакты' },
  { href: AppPath.Support, label: 'Поддержка' },
];

export const FOOTER_COPYRIGHT = '© 2026, ООО ПО «Волжский инструмент». Все права защищены.';

export const FOOTER_VERSION = 'Прототип ВИ-Портала • TZ v1.1';
