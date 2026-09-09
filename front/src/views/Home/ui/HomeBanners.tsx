'use client';

import { ScrollShadow } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useEffect, useMemo, useRef } from 'react';

import type { BannersSettings } from '@/core/shared/api/content';

import styles from './HomeBanners.module.css';

type HomeBannersProps = {
  banners: BannersSettings;
};

export const HomeBanners = ({ banners }: HomeBannersProps): JSX.Element | null => {
  const slides = useMemo(
    () =>
      [...(banners.items ?? [])]
        .filter((item) => item.is_active && Boolean(item.image))
        .sort((a, b) => a.sort_order - b.sort_order || a.id.localeCompare(b.id)),
    [banners.items],
  );

  const scrollerRef = useRef<HTMLDivElement>(null);
  const delayMs = Math.max(1, banners.delay_sec || 6) * 1000;

  useEffect(() => {
    if (slides.length < 2) {
      return;
    }

    const timer = window.setInterval(() => {
      const node = scrollerRef.current;

      if (!node) {
        return;
      }

      const amount = Math.min(node.clientWidth * 0.85, 380);
      const atEnd = node.scrollLeft + node.clientWidth >= node.scrollWidth - 8;

      if (atEnd) {
        node.scrollTo({ behavior: 'smooth', left: 0 });
        return;
      }

      node.scrollBy({ behavior: 'smooth', left: amount });
    }, delayMs);

    return () => window.clearInterval(timer);
  }, [delayMs, slides.length]);

  if (slides.length === 0) {
    return null;
  }

  return (
    <section aria-label="Баннеры" className={clsx(styles.root)}>
      <div className={clsx(styles.container)}>
        <ScrollShadow
          className={clsx(styles.scrollShadow)}
          hideScrollBar
          orientation="horizontal"
          ref={scrollerRef}
          size={48}
        >
          <div className={clsx(styles.scroller)}>
            {slides.map((slide) => {
              const href = slide.link?.trim() || '';
              const body = (
                <>
                  <div
                    className={clsx(styles.image)}
                    style={{ backgroundImage: `url(${slide.image})` }}
                  />
                  <div className={clsx(styles.overlay)} />
                  <div className={clsx(styles.copy)}>
                    {slide.title ? <h3 className={clsx(styles.title)}>{slide.title}</h3> : null}
                    {slide.subtitle ? (
                      <p className={clsx(styles.subtitle)}>{slide.subtitle}</p>
                    ) : null}
                  </div>
                </>
              );

              if (href) {
                return (
                  <Link
                    className={clsx(styles.card)}
                    href={href}
                    key={slide.id}
                    rel="noopener noreferrer"
                    target="_blank"
                  >
                    {body}
                  </Link>
                );
              }

              return (
                <div className={clsx(styles.card)} key={slide.id}>
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
