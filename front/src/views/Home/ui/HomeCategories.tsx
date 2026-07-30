'use client';

import clsx from 'clsx';
import Link from 'next/link';

import { formatPositionsCount } from '@/core/shared/lib/pluralize';

import type { HomeCategoriesContent } from '../lib/types';

import styles from './HomeCategories.module.css';

type HomeCategoriesProps = {
  content: HomeCategoriesContent;
};

export const HomeCategories = ({ content }: HomeCategoriesProps): JSX.Element => {
  return (
    <section className={clsx(styles.root)}>
      <div className={clsx(styles.container)}>
        <div className={clsx(styles.header)}>
          <div className={clsx(styles.heading)}>
            <span className={clsx(styles.eyebrow)}>{content.eyebrow}</span>
            <h2 className={clsx(styles.title)}>{content.title}</h2>
          </div>
          <Link className={clsx(styles.viewAll)} href={content.viewAllHref}>
            {content.viewAllLabel}
          </Link>
        </div>

        <div className={clsx(styles.grid)}>
          {content.items.map((item) => (
            <Link className={clsx(styles.card)} href={item.href} key={item.id}>
              <div
                className={clsx(styles.media, !item.imageUrl && styles.mediaPlaceholder)}
                style={item.imageUrl ? { backgroundImage: `url(${item.imageUrl})` } : undefined}
              />
              <div className={clsx(styles.body)}>
                <h3 className={clsx(styles.name)}>{item.name}</h3>
                <span className={clsx(styles.count)}>
                  {formatPositionsCount(item.positionsCount)}
                </span>
                <span className={clsx(styles.look)}>Смотреть →</span>
              </div>
            </Link>
          ))}
        </div>
      </div>
    </section>
  );
};
