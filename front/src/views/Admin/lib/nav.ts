import { AppPath, getAdminContentPath } from '@/core/shared/router/paths';

export type AdminNavGroup = {
  items: AdminNavItem[];
  title?: string;
};

export type AdminNavItem = {
  href: string;
  label: string;
};

export const ADMIN_NAV: AdminNavGroup[] = [
  {
    items: [{ href: AppPath.Admin, label: 'Сводка' }],
  },
  {
    items: [
      { href: getAdminContentPath('home'), label: 'Главная страница' },
      { href: AppPath.AdminBanners, label: 'Баннеры' },
      { href: getAdminContentPath('about'), label: 'О компании' },
      { href: getAdminContentPath('terms'), label: 'Условия работы' },
      { href: getAdminContentPath('promo'), label: 'Акции' },
      { href: getAdminContentPath('certificates'), label: 'Сертификаты' },
      { href: getAdminContentPath('contacts'), label: 'Контакты' },
    ],
    title: 'Контент',
  },
  {
    items: [
      { href: AppPath.AdminProducts, label: 'Товары' },
      { href: AppPath.AdminCategories, label: 'Категории и бренды' },
    ],
    title: 'Каталог',
  },
  {
    items: [{ href: AppPath.AdminLegal, label: 'Документы и соглашения' }],
    title: 'Юридические документы',
  },
  {
    items: [
      { href: AppPath.AdminUsers, label: 'Пользователи портала' },
      { href: AppPath.AdminSignupRequests, label: 'Заявки на регистрацию' },
      { href: AppPath.AdminFeedback, label: 'Отзывы по заказам' },
      { href: AppPath.AdminSupport, label: 'Обращения поддержки' },
    ],
    title: 'Пользователи',
  },
  {
    items: [{ href: AppPath.AdminAuditLog, label: 'Журнал действий' }],
    title: 'Журнал',
  },
];

export const formatAdminDateTime = (value: string): string => {
  if (!value) {
    return '—';
  }

  const date = new Date(value);

  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat('ru-RU', {
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    month: '2-digit',
    year: 'numeric',
  }).format(date);
};
