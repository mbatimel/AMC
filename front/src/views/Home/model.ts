import { combine } from 'effector';

import type { BannersSettings, ContentPages } from '@/core/shared/api/content';

import { $banners, $content } from '@/core/entities/content';
import { AppPath } from '@/core/shared/router/paths';

import type { HomePageContent, HomePromoCard } from './lib/types';

import { HOME_PAGE_MOCK } from './lib/mocks';

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

const toHomeContent = (
  content: ContentPages | null,
  banners: BannersSettings | null,
): HomePageContent => {
  if (!content) {
    return HOME_PAGE_MOCK;
  }

  const home = content.home;
  const promoCards = banners ? toPromoCards(banners) : [];

  return {
    categories: HOME_PAGE_MOCK.categories,
    hero: {
      ...HOME_PAGE_MOCK.hero,
      bullets: home.features.length > 0 ? home.features : HOME_PAGE_MOCK.hero.bullets,
      description: home.hero_subtitle || HOME_PAGE_MOCK.hero.description,
      primaryCta: {
        href: AppPath.Catalog,
        label: home.hero_button || HOME_PAGE_MOCK.hero.primaryCta.label,
      },
      secondaryCta: { href: AppPath.Terms, label: 'Условия для оптовиков' },
      stats: home.stats.length > 0 ? home.stats : HOME_PAGE_MOCK.hero.stats,
      title: home.hero_title || HOME_PAGE_MOCK.hero.title,
    },
    promos: {
      items: promoCards.length > 0 ? promoCards : HOME_PAGE_MOCK.promos.items,
      title: content.promo.title || HOME_PAGE_MOCK.promos.title,
    },
  };
};

/**
 * Контент главной: приходит из редактора админки (`/admin/content/home`
 * и `/admin/banners`), при недоступности API — статичный fallback.
 */
export const $homeContent = combine($content, $banners, toHomeContent);
