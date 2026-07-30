'use client';

import clsx from 'clsx';
import { useUnit } from 'effector-react';

import styles from '../Home.module.css';
import { $homeContent } from '../model';
import { HomeCategories } from './HomeCategories';
import { HomeHero } from './HomeHero';
import { HomePromos } from './HomePromos';

export const HomeContent = (): JSX.Element => {
  const content = useUnit($homeContent);

  return (
    <div className={clsx(styles.root)}>
      <HomeHero content={content.hero} />
      <HomePromos content={content.promos} />
      <HomeCategories content={content.categories} />
    </div>
  );
};
