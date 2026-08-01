import { combine, createEffect, createEvent, createStore, sample } from 'effector';

import type { BannersSettings, ContentPages } from '@/core/shared/api/content';
import type { Category } from '@/core/shared/api/products';

import { $banners, $content } from '@/core/entities/content';
import { listCategoriesRequest, listProductsRequest } from '@/core/shared/api/products';
import { AppPath } from '@/core/shared/router/paths';

import type { HomeCategoryCard, HomePageContent, HomePromoCard } from './lib/types';

import { catalogHref, HOME_PAGE_MOCK } from './lib/mocks';

const maxHomeCategories = 6;

const toPromoCards = (banners: BannersSettings): HomePromoCard[] =>
  banners.items
    .filter((item) => item.is_active)
    .map((item, index) => ({
      href: item.link || AppPath.Catalog,
      id: item.id,
      text: item.subtitle,
      title: item.title,
      tone: index % 2 === 0 ? ('red' as const) : ('dark' as const),
    }));

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

export const fetchHomeCategoriesFx = createEffect(async (): Promise<HomeCategoryCard[]> => {
  const categories = await listCategoriesRequest();

  return Promise.all(
    categories
      .filter((category) => !category.parent_id)
      .slice(0, maxHomeCategories)
      .map(toCategoryCard),
  );
});

/** Каталог реальных категорий с главной; до загрузки — статичный fallback. */
export const $homeCategories = createStore<HomeCategoryCard[]>(HOME_PAGE_MOCK.categories.items).on(
  fetchHomeCategoriesFx.doneData,
  (_, items) => items,
);

const $isHomeCategoriesLoaded = createStore(false).on(fetchHomeCategoriesFx.done, () => true);

/* eslint-disable perfectionist/sort-objects -- effector sample: clock -> source -> filter -> target */
sample({
  clock: homeCategoriesRequested,
  source: { loaded: $isHomeCategoriesLoaded, pending: fetchHomeCategoriesFx.pending },
  filter: ({ loaded, pending }) => !loaded && !pending,
  target: fetchHomeCategoriesFx,
});
/* eslint-enable perfectionist/sort-objects */

const toHomeContent = (
  content: ContentPages | null,
  banners: BannersSettings | null,
  categories: HomeCategoryCard[],
): HomePageContent => {
  const home = content?.home;
  const promoCards = banners ? toPromoCards(banners) : [];

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
      items: promoCards.length > 0 ? promoCards : HOME_PAGE_MOCK.promos.items,
      title: content?.promo.title || HOME_PAGE_MOCK.promos.title,
    },
  };
};

/**
 * Контент главной: приходит из редактора админки (admin-front, разделы
 * «Главная страница» и «Баннеры») и из реального каталога (категории),
 * при недоступности API — статичный fallback.
 */
export const $homeContent = combine($content, $banners, $homeCategories, toHomeContent);
