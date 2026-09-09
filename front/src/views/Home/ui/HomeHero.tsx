'use client';

import { Button } from '@heroui/react';
import clsx from 'clsx';
import Link from 'next/link';
import { useRouter } from 'next/navigation';

import { HtmlContent } from '@/core/shared/ui/HtmlContent';

import type { HomeHeroContent } from '../lib/types';

import styles from './HomeHero.module.css';

type HomeHeroProps = {
  content: HomeHeroContent;
  isPending?: boolean;
};

export const HomeHero = ({ content, isPending = false }: HomeHeroProps): JSX.Element => {
  const router = useRouter();
  const { badge, bullets, description, imageUrl, primaryCta, secondaryCta, stats, title } = content;

  if (isPending) {
    return (
      <section aria-busy="true" className={clsx(styles.root)} role="status">
        <div className={clsx(styles.media, styles.mediaPlaceholder)} />
        <div className={clsx(styles.overlay)} />
        <div className={clsx(styles.inner)}>
          <div className={clsx(styles.copy)}>
            <span className={clsx(styles.skeletonLine, styles.skeletonBadge)} />
            <span className={clsx(styles.skeletonLine, styles.skeletonTitle)} />
            <span className={clsx(styles.skeletonLine, styles.skeletonText)} />
            <span className={clsx(styles.skeletonLine, styles.skeletonTextShort)} />
          </div>
        </div>
      </section>
    );
  }

  return (
    <section className={clsx(styles.root)}>
      <div
        className={clsx(styles.media, !imageUrl && styles.mediaPlaceholder)}
        style={imageUrl ? { backgroundImage: `url(${imageUrl})` } : undefined}
      />
      <div className={clsx(styles.overlay)} />

      <div className={clsx(styles.inner)}>
        <div className={clsx(styles.copy)}>
          {badge ? <span className={clsx(styles.badge)}>{badge}</span> : null}
          {title ? <h1 className={clsx(styles.title)}>{title}</h1> : null}
          {description ? (
            <HtmlContent className={clsx(styles.description)} text={description} />
          ) : null}

          {bullets.length > 0 ? (
            <ul className={clsx(styles.bullets)}>
              {bullets.map((bullet) => (
                <li key={bullet}>{bullet}</li>
              ))}
            </ul>
          ) : null}

          {primaryCta.label || secondaryCta.label ? (
            <div className={clsx(styles.actions)}>
              {primaryCta.label ? (
                <Button
                  className={clsx(styles.primaryCta)}
                  onPress={() => router.push(primaryCta.href)}
                  variant="primary"
                >
                  {primaryCta.label}
                </Button>
              ) : null}
              {secondaryCta.label ? (
                secondaryCta.href === '#' ? (
                  <span className={clsx(styles.secondaryCta)}>{secondaryCta.label}</span>
                ) : (
                  <Link className={clsx(styles.secondaryCta)} href={secondaryCta.href}>
                    {secondaryCta.label}
                  </Link>
                )
              ) : null}
            </div>
          ) : null}
        </div>
      </div>

      {stats.length > 0 ? (
        <div className={clsx(styles.stats)}>
          <div className={clsx(styles.statsInner)}>
            {stats.map((stat) => (
              <div className={clsx(styles.statItem)} key={stat.label}>
                <strong className={clsx(styles.statValue)}>{stat.value}</strong>
                <span className={clsx(styles.statLabel)}>{stat.label}</span>
              </div>
            ))}
          </div>
        </div>
      ) : null}
    </section>
  );
};
