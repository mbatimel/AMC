import { AppPath } from '@/core/shared/router/paths';

import type { HomePageContent } from './types';

const catalogHref = (categoryID?: string): string => {
  if (!categoryID) {
    return AppPath.Catalog;
  }

  return `${AppPath.Catalog}?categoryID=${encodeURIComponent(categoryID)}`;
};

/** Fallback / mock до CMS из админки. */
export const HOME_PAGE_MOCK: HomePageContent = {
  categories: {
    eyebrow: 'Каталог',
    items: [
      {
        href: catalogHref('taps'),
        id: 'taps',
        imageUrl: null,
        name: 'Метчики',
        positionsCount: 3,
      },
      {
        href: catalogHref('drills'),
        id: 'drills',
        imageUrl: null,
        name: 'Сверла',
        positionsCount: 3,
      },
      {
        href: catalogHref('dies'),
        id: 'dies',
        imageUrl: null,
        name: 'Плашки',
        positionsCount: 3,
      },
      {
        href: catalogHref('files'),
        id: 'files',
        imageUrl: null,
        name: 'Напильники',
        positionsCount: 3,
      },
      {
        href: catalogHref('wrenches'),
        id: 'wrenches',
        imageUrl: null,
        name: 'Ключи',
        positionsCount: 3,
      },
      {
        href: catalogHref('garden'),
        id: 'garden',
        imageUrl: null,
        name: 'Садовый инструмент',
        positionsCount: 3,
      },
    ],
    title: 'Популярные категории',
    viewAllHref: AppPath.Catalog,
    viewAllLabel: 'Весь каталог',
  },
  hero: {
    badge: 'Производство с 1993 года',
    bullets: [
      'Более 4 200 позиций в наличии на складе',
      'Индивидуальные цены для оптовых клиентов',
      'Отгрузка по всей России и СНГ',
    ],
    description:
      'Собственное производство металлорежущего инструмента, складская программа и быстрая отгрузка для оптовых клиентов.',
    imageUrl: null,
    primaryCta: {
      href: AppPath.Catalog,
      label: 'Перейти в каталог',
    },
    secondaryCta: {
      href: '#',
      label: 'Условия для оптовиков',
    },
    stats: [
      { label: 'SKU на складе', value: '4 200+' },
      { label: 'оптовых клиентов', value: '820+' },
      { label: 'лет на рынке', value: '33' },
      { label: 'заказов вовремя', value: '99.2%' },
    ],
    title: 'Профессиональный инструмент оптом',
  },
  promos: {
    items: [
      {
        href: AppPath.Catalog,
        id: 'promo-delivery',
        text: 'При заказе от 50 000 ₽ — бесплатная доставка по городу в течение 24 часов.',
        title: 'Доставка по Самаре',
        tone: 'red',
      },
      {
        href: AppPath.Catalog,
        id: 'promo-dies',
        text: 'Спеццена на плашки 9ХС для постоянных клиентов до конца месяца.',
        title: 'Плашки 9ХС',
        tone: 'dark',
      },
      {
        href: AppPath.Catalog,
        id: 'promo-drills',
        text: 'Линейка свёрл VI-Pro со скидкой при заказе упаковками.',
        title: 'Свёрла VI-Pro',
        tone: 'red',
      },
    ],
    title: 'Акции и спецпредложения',
  },
};
