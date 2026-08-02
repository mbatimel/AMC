'use client';

import { Button, ScrollShadow } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useRef } from 'react';

import type { HomePromosContent } from '../lib/types';

import styles from './HomePromos.module.css';

type HomePromosProps = {
  content: HomePromosContent;
};

export const HomePromos = ({ content }: HomePromosProps): JSX.Element | null => {
  const scrollerRef = useRef<HTMLDivElement>(null);

  if (content.items.length === 0) {
    return null;
  }

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
            <Button
              aria-label="Предыдущие акции"
              className={clsx(styles.navButton)}
              isIconOnly
              onPress={() => scrollByCard(-1)}
              size="sm"
              variant="secondary"
            >
              ←
            </Button>
            <Button
              aria-label="Следующие акции"
              className={clsx(styles.navButton)}
              isIconOnly
              onPress={() => scrollByCard(1)}
              size="sm"
              variant="secondary"
            >
              →
            </Button>
          </div>
        </div>

        <ScrollShadow
          className={clsx(styles.scrollShadow)}
          hideScrollBar
          orientation="horizontal"
          ref={scrollerRef}
          size={48}
        >
          <div className={clsx(styles.scroller)}>
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
        </ScrollShadow>
      </div>
    </section>
  );
};
