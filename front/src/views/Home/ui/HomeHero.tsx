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
};

export const HomeHero = ({ content }: HomeHeroProps): JSX.Element => {
  const router = useRouter();
  const { badge, bullets, description, imageUrl, primaryCta, secondaryCta, stats, title } = content;

  return (
    <section className={clsx(styles.root)}>
      <div
        className={clsx(styles.media, !imageUrl && styles.mediaPlaceholder)}
        style={imageUrl ? { backgroundImage: `url(${imageUrl})` } : undefined}
      />
      <div className={clsx(styles.overlay)} />

      <div className={clsx(styles.inner)}>
        <div className={clsx(styles.copy)}>
          <span className={clsx(styles.badge)}>{badge}</span>
          <h1 className={clsx(styles.title)}>{title}</h1>
          <HtmlContent className={clsx(styles.description)} text={description} />

          <ul className={clsx(styles.bullets)}>
            {bullets.map((bullet) => (
              <li key={bullet}>{bullet}</li>
            ))}
          </ul>

          <div className={clsx(styles.actions)}>
            <Button
              className={clsx(styles.primaryCta)}
              onPress={() => router.push(primaryCta.href)}
              variant="primary"
            >
              {primaryCta.label}
            </Button>
            {secondaryCta.href === '#' ? (
              <span className={clsx(styles.secondaryCta)}>{secondaryCta.label}</span>
            ) : (
              <Link className={clsx(styles.secondaryCta)} href={secondaryCta.href}>
                {secondaryCta.label}
              </Link>
            )}
          </div>
        </div>
      </div>

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
    </section>
  );
};
