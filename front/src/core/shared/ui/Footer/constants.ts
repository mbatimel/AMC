import { AppPath } from '@/core/shared/router/paths';

export type FooterLinkItem = {
  href: string;
  label: string;
};

export const FOOTER_BRANDS = [
  'Волжский инструмент',
  'ВИЗ',
  'Krotec',
  'Schwert',
  'Канаш',
  'Delta',
  'Белый Медведь',
  'BOON',
  'Камышин',
  'Сервис Ключ',
  'СТиЗ',
];

export const FOOTER_BRANDS_TITLE = 'Бренды на портале';

export const FOOTER_DESCRIPTION =
  'Производитель и оптовый поставщик профессионального инструмента.';

export const FOOTER_CATALOG_LINKS: FooterLinkItem[] = [
  { href: '#', label: 'Металлорежущий' },
  { href: '#', label: 'Метчики' },
  { href: '#', label: 'Сверла' },
  { href: '#', label: 'Плашки' },
  { href: '#', label: 'Слесарный' },
  { href: '#', label: 'Садовый' },
];

export const FOOTER_CABINET_LINKS: FooterLinkItem[] = [
  { href: AppPath.Cabinet, label: 'Личный кабинет' },
  { href: '#', label: 'Мои заказы' },
  { href: '#', label: 'Документы' },
  { href: '#', label: 'Избранное' },
  { href: '#', label: 'Профиль' },
];

export const FOOTER_INFO_LINKS: FooterLinkItem[] = [
  { href: '#', label: 'О компании' },
  { href: '#', label: 'Условия работы' },
  { href: '#', label: 'Сертификаты' },
  { href: '#', label: 'Контакты' },
  { href: '#', label: 'Поддержка' },
];

export const FOOTER_LEGAL_LINKS: FooterLinkItem[] = [
  { href: '#', label: 'Оферта' },
  { href: '#', label: 'Конфиденциальность' },
  { href: '#', label: 'Согласие на ПД' },
  { href: '#', label: 'Пользовательское соглашение' },
];

export const FOOTER_COPYRIGHT = '© 2026, ООО ПО «Волжский инструмент». Все права защищены.';

export const FOOTER_VERSION = 'Прототип ВИ-Портала • TZ v1.1';
