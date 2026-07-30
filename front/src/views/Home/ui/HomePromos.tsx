'use client';

import clsx from 'clsx';
import Link from 'next/link';
import { useRef } from 'react';

import type { HomePromosContent } from '../lib/types';

import styles from './HomePromos.module.css';

type HomePromosProps = {
  content: HomePromosContent;
};

export const HomePromos = ({ content }: HomePromosProps): JSX.Element => {
  const scrollerRef = useRef<HTMLDivElement>(null);

  const scrollByCard = (direction: -1 | 1): void => {
    const node = scrollerRef.current;

    if (!node) {
      return;
    }

    const amount = Math.min(node.clientWidth * 0.85, 380);

    node.scrollBy({ behavior: 'smooth', left: amount * direction });
  };

  return (
    <section className={clsx(styles.root)}>
      <div className={clsx(styles.container)}>
        <div className={clsx(styles.header)}>
          <h2 className={clsx(styles.title)}>{content.title}</h2>
          <div className={clsx(styles.nav)}>
            <button
              aria-label="Предыдущие акции"
              className={clsx(styles.navButton)}
              onClick={() => scrollByCard(-1)}
              type="button"
            >
              ←
            </button>
            <button
              aria-label="Следующие акции"
              className={clsx(styles.navButton)}
              onClick={() => scrollByCard(1)}
              type="button"
            >
              →
            </button>
          </div>
        </div>

        <div className={clsx(styles.scroller)} ref={scrollerRef}>
          {content.items.map((item) => {
            const cardClass = clsx(
              styles.card,
              item.tone === 'red' ? styles.cardRed : styles.cardDark,
            );

            const body = (
              <>
                <h3 className={clsx(styles.cardTitle)}>{item.title}</h3>
                <p className={clsx(styles.cardText)}>{item.text}</p>
              </>
            );

            if (item.href) {
              return (
                <Link className={cardClass} href={item.href} key={item.id}>
                  {body}
                </Link>
              );
            }

            return (
              <div className={cardClass} key={item.id}>
                {body}
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
};
