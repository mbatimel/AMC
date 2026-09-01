import { AppPath, getContentPath } from '@/core/shared/router/paths';

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
    items: [{ href: AppPath.Home, label: 'Сводка' }],
  },
  {
    items: [
      { href: getContentPath('home'), label: 'Главная страница' },
      { href: AppPath.Banners, label: 'Баннеры' },
      { href: getContentPath('about'), label: 'О компании' },
      { href: getContentPath('terms'), label: 'Условия работы' },
      { href: AppPath.Certificates, label: 'Сертификаты' },
      { href: getContentPath('contacts'), label: 'Контакты' },
    ],
    title: 'Контент',
  },
  {
    items: [
      { href: AppPath.Products, label: 'Товары' },
      { href: AppPath.Categories, label: 'Категории и бренды' },
      { href: AppPath.Promotions, label: 'Акции' },
    ],
    title: 'Каталог',
  },
  {
    items: [{ href: AppPath.Legal, label: 'Документы и соглашения' }],
    title: 'Юридические документы',
  },
  {
    items: [
      { href: AppPath.Users, label: 'Пользователи портала' },
      { href: AppPath.SignupRequests, label: 'Заявки на регистрацию' },
      { href: AppPath.Feedback, label: 'Отзывы по заказам' },
      { href: AppPath.Support, label: 'Обращения поддержки' },
    ],
    title: 'Пользователи',
  },
  {
    items: [{ href: AppPath.AuditLog, label: 'Журнал действий' }],
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
