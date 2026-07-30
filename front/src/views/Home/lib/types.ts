export type HomeCategoriesContent = {
  eyebrow: string;
  items: HomeCategoryCard[];
  title: string;
  viewAllHref: string;
  viewAllLabel: string;
};

export type HomeCategoryCard = {
  href: string;
  id: string;
  imageUrl: null | string;
  name: string;
  positionsCount: number;
};

export type HomeHeroContent = {
  badge: string;
  bullets: string[];
  description: string;
  imageUrl: null | string;
  primaryCta: { href: string; label: string };
  secondaryCta: { href: string; label: string };
  stats: HomeStatItem[];
  title: string;
};

export type HomePageContent = {
  categories: HomeCategoriesContent;
  hero: HomeHeroContent;
  promos: HomePromosContent;
};

export type HomePromoCard = {
  href?: string;
  id: string;
  text: string;
  title: string;
  tone: HomePromoTone;
};

export type HomePromosContent = {
  items: HomePromoCard[];
  title: string;
};

export type HomePromoTone = 'dark' | 'red';

export type HomeStatItem = {
  label: string;
  value: string;
};
