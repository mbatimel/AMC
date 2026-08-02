import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { ContentPages } from '@/core/shared/api/content';
import type { Category } from '@/core/shared/api/products';
import type { Promotion } from '@/core/shared/api/promotions';

import { $content } from '@/core/entities/content';
import { listCategoriesRequest, listProductsRequest } from '@/core/shared/api/products';
import { listPromotionsRequest } from '@/core/shared/api/promotions';
import { AppPath, getCatalogPromotionPath } from '@/core/shared/router/paths';

import type { HomeCategoryCard, HomePageContent, HomePromoCard } from './lib/types';

import { catalogHref, HOME_PAGE_MOCK } from './lib/mocks';

const maxHomeCategories = 6;

const formatPromoPeriod = (startsAt: string, endsAt: string): string => {
  const start = new Date(startsAt);
  const end = new Date(endsAt);

  if (Number.isNaN(start.getTime()) || Number.isNaN(end.getTime())) {
    return '';
  }

  const format = (date: Date): string =>
    date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' });

  return `${format(start)} – ${format(end)}`;
};

const toPromoCards = (promotions: Promotion[]): HomePromoCard[] =>
  promotions.map((promo, index) => {
    const period = formatPromoPeriod(promo.starts_at, promo.ends_at);
    const discount = `−${promo.discount_percent}%`;

    return {
      href: getCatalogPromotionPath(promo.id, promo.name),
      id: promo.id,
      text: period ? `${discount} · ${period}` : discount,
      title: promo.name,
      tone: index % 2 === 0 ? ('red' as const) : ('dark' as const),
    };
  });

const toCategoryCard = async (category: Category): Promise<HomeCategoryCard> => {
  const { pagination } = await listProductsRequest({ categoryID: category.id, limit: 1 });

  return {
    href: catalogHref(category.id),
    id: category.id,
    imageUrl: null,
    name: category.name,
    positionsCount: pagination.total,
  };
};

export const homeCategoriesRequested = createEvent();
export const homePromosRequested = createEvent();

export const fetchHomeCategoriesFx = createEffect(async (): Promise<HomeCategoryCard[]> => {
  const categories = await listCategoriesRequest();

  return Promise.all(
    categories
      .filter((category) => !category.parent_id)
      .slice(0, maxHomeCategories)
      .map(toCategoryCard),
  );
});

export const fetchHomePromosFx = createEffect(async (): Promise<HomePromoCard[]> => {
  const promotions = await listPromotionsRequest();

  return toPromoCards(promotions.filter((promo) => promo.status === 'active'));
});

/** Каталог реальных категорий с главной; до загрузки — статичный fallback. */
export const $homeCategories = createStore<HomeCategoryCard[]>(HOME_PAGE_MOCK.categories.items).on(
  fetchHomeCategoriesFx.doneData,
  (_, items) => items,
);

export const $homePromos = createStore<HomePromoCard[]>([]).on(
  fetchHomePromosFx.doneData,
  (_, items) => items,
);

const $isHomeCategoriesLoaded = createStore(false).on(fetchHomeCategoriesFx.done, () => true);
const $isHomePromosLoaded = createStore(false).on(fetchHomePromosFx.done, () => true);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> target */
sample({
  clock: homeCategoriesRequested,
  source: { loaded: $isHomeCategoriesLoaded, pending: fetchHomeCategoriesFx.pending },
  filter: ({ loaded, pending }) => !loaded && !pending,
  target: fetchHomeCategoriesFx,
});

sample({
  clock: homePromosRequested,
  source: { loaded: $isHomePromosLoaded, pending: fetchHomePromosFx.pending },
  filter: ({ loaded, pending }) => !loaded && !pending,
  target: fetchHomePromosFx,
});
/* eslint-enable perfectionist/sort-objects */

const toHomeContent = (
  content: ContentPages | null,
  categories: HomeCategoryCard[],
  promoCards: HomePromoCard[],
): HomePageContent => {
  const home = content?.home;

  return {
    categories: { ...HOME_PAGE_MOCK.categories, items: categories },
    hero: {
      ...HOME_PAGE_MOCK.hero,
      bullets: home && home.features.length > 0 ? home.features : HOME_PAGE_MOCK.hero.bullets,
      description: home?.hero_subtitle || HOME_PAGE_MOCK.hero.description,
      primaryCta: {
        href: AppPath.Catalog,
        label: home?.hero_button || HOME_PAGE_MOCK.hero.primaryCta.label,
      },
      secondaryCta: { href: AppPath.Terms, label: 'Условия для оптовиков' },
      stats: home && home.stats.length > 0 ? home.stats : HOME_PAGE_MOCK.hero.stats,
      title: home?.hero_title || HOME_PAGE_MOCK.hero.title,
    },
    promos: {
      items: promoCards,
      title: 'Акции и спецпредложения',
    },
  };
};

/**
 * Контент главной: редактор админки (home), реальные акции и каталог категорий.
 */
export const $homeContent = combine($content, $homeCategories, $homePromos, toHomeContent);
