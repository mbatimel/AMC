'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';
import { useEffect } from 'react';

import { useContent } from '@/core/entities/content';

import styles from '../Home.module.css';
import { $homeContent, homeCategoriesRequested } from '../model';
import { HomeCategories } from './HomeCategories';
import { HomeHero } from './HomeHero';
import { HomePromos } from './HomePromos';

export const HomeContent = (): JSX.Element => {
  const [content, requestCategories] = useUnit([$homeContent, homeCategoriesRequested]);

  useContent();

  useEffect(() => {
    requestCategories();
  }, [requestCategories]);

  return (
    <div className={clsx(styles.root)}>
      <HomeHero content={content.hero} />
      <HomePromos content={content.promos} />
      <HomeCategories content={content.categories} />
    </div>
  );
};
