'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { useContent } from '@/core/entities/content';

import styles from '../Home.module.css';
import { $homeContent, homeCategoriesRequested, homePromosRequested } from '../model';
import { HomeBanners } from './HomeBanners';
import { HomeCategories } from './HomeCategories';
import { HomeHero } from './HomeHero';
import { HomePromos } from './HomePromos';

export const HomeContent = (): JSX.Element => {
  const [content, requestCategories, requestPromos] = useUnit([
    $homeContent,
    homeCategoriesRequested,
    homePromosRequested,
  ]);

  const { banners, content: portalContent, isPending } = useContent();
  const isHeroPending = isPending && !portalContent?.home;

  useEffect(() => {
    requestCategories();
    requestPromos();
  }, [requestCategories, requestPromos]);

  return (
    <div className={clsx(styles.root)}>
      <HomeHero content={content.hero} isPending={isHeroPending} />
      {banners ? <HomeBanners banners={banners} /> : null}
      <HomePromos content={content.promos} />
      <HomeCategories content={content.categories} />
    </div>
  );
};
