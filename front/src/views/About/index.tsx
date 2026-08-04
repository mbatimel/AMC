'use client';

import clsx from 'clsx';
import Link from 'next/link';

import { useContent } from '@/core/entities/content';
import { IconLocation, IconPhone } from '@/core/shared/icons';
import { AppPath } from '@/core/shared/router/paths';
import { HEADER_PHONE_MAIN } from '@/core/shared/ui/Header/constants';
import { Page } from '@/core/shared/ui/Page';

import styles from './About.module.css';
import {
  ABOUT_CTA_BADGE,
  ABOUT_CTA_TEXT,
  ABOUT_CTA_TITLE,
  ABOUT_DIRECTIONS,
  ABOUT_DIRECTIONS_BADGE,
  ABOUT_DIRECTIONS_SUBTITLE,
  ABOUT_DIRECTIONS_TITLE,
  ABOUT_HERO_BADGE,
  ABOUT_HERO_SUBTITLE,
  ABOUT_OFFICES,
  ABOUT_OFFICES_BADGE,
  ABOUT_OFFICES_SUBTITLE,
  ABOUT_OFFICES_TITLE,
  ABOUT_PROFILE_BADGE,
  ABOUT_PROFILE_TITLE,
} from './lib/aboutData';

const toTelHref = (phone: string): string => `tel:${phone.replace(/[^\d+]/g, '')}`;

const splitParagraphs = (text: string): string[] =>
  text
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean);

export const About = (): JSX.Element => {
  const { content, error, isPending } = useContent();
  const about = content?.about;
  const title = about?.title ?? 'О компании';
  const paragraphs = about?.text ? splitParagraphs(about.text) : [];

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <section className={clsx(styles.hero)}>
          <div className={clsx(styles.heroInner)}>
            <p className={clsx(styles.heroBadge)}>{ABOUT_HERO_BADGE}</p>
            <h1 className={clsx(styles.heroTitle)}>{title}</h1>
            <p className={clsx(styles.heroDescription)}>{ABOUT_HERO_SUBTITLE}</p>
          </div>
        </section>

        <div className={clsx(styles.container)}>
          <section aria-labelledby="about-profile-title" className={clsx(styles.profile)}>
            <div className={clsx(styles.sectionIntro)}>
              <p className={clsx(styles.sectionBadge)}>{ABOUT_PROFILE_BADGE}</p>
              <h2 className={clsx(styles.sectionTitle)} id="about-profile-title">
                {ABOUT_PROFILE_TITLE}
              </h2>
            </div>

            {isPending && !about ? <p className={clsx(styles.status)}>Загрузка…</p> : null}
            {error && !about ? <p className={clsx(styles.error)}>{error}</p> : null}

            {paragraphs.length > 0 ? (
              <div className={clsx(styles.profileText)}>
                {paragraphs.map((paragraph) => (
                  <p key={paragraph.slice(0, 24)}>{paragraph}</p>
                ))}
              </div>
            ) : null}
          </section>

          <section aria-labelledby="about-directions-title" className={clsx(styles.section)}>
            <div className={clsx(styles.sectionIntro)}>
              <p className={clsx(styles.sectionBadge)}>{ABOUT_DIRECTIONS_BADGE}</p>
              <h2 className={clsx(styles.sectionTitle)} id="about-directions-title">
                {ABOUT_DIRECTIONS_TITLE}
              </h2>
              <p className={clsx(styles.sectionSubtitle)}>{ABOUT_DIRECTIONS_SUBTITLE}</p>
            </div>

            <div className={clsx(styles.directionsGrid)}>
              {ABOUT_DIRECTIONS.map((direction) => (
                <article className={clsx(styles.directionCard)} key={direction.title}>
                  <span aria-hidden className={clsx(styles.directionIcon)}>
                    <direction.Icon currentColor="currentColor" height={20} width={20} />
                  </span>
                  <h3 className={clsx(styles.directionTitle)}>{direction.title}</h3>
                  <p className={clsx(styles.directionText)}>{direction.description}</p>
                </article>
              ))}
            </div>
          </section>

          <section aria-labelledby="about-offices-title" className={clsx(styles.section)}>
            <div className={clsx(styles.sectionIntro)}>
              <p className={clsx(styles.sectionBadge)}>{ABOUT_OFFICES_BADGE}</p>
              <h2 className={clsx(styles.sectionTitle)} id="about-offices-title">
                {ABOUT_OFFICES_TITLE}
              </h2>
              <p className={clsx(styles.sectionSubtitle)}>{ABOUT_OFFICES_SUBTITLE}</p>
            </div>

            <div className={clsx(styles.officesGrid)}>
              {ABOUT_OFFICES.map((office) => (
                <article className={clsx(styles.officeCard)} key={office.city}>
                  <div className={clsx(styles.officeHeader)}>
                    <span aria-hidden className={clsx(styles.officePin)}>
                      <IconLocation currentColor="currentColor" height={16} width={16} />
                    </span>
                    <h3 className={clsx(styles.officeCity)}>{office.city}</h3>
                    {office.isMain ? (
                      <span className={clsx(styles.officeMain)}>главный офис</span>
                    ) : null}
                  </div>
                  <p className={clsx(styles.officeText)}>{office.description}</p>
                </article>
              ))}
            </div>
          </section>

          <section aria-labelledby="about-cta-title" className={clsx(styles.cta)}>
            <div className={clsx(styles.ctaCopy)}>
              <p className={clsx(styles.sectionBadge)}>{ABOUT_CTA_BADGE}</p>
              <h2 className={clsx(styles.ctaTitle)} id="about-cta-title">
                {ABOUT_CTA_TITLE}
              </h2>
              <p className={clsx(styles.ctaText)}>{ABOUT_CTA_TEXT}</p>
            </div>
            <div className={clsx(styles.ctaActions)}>
              <Link className={clsx(styles.primaryButton)} href={AppPath.Support}>
                Отправить заявку
              </Link>
              <a className={clsx(styles.secondaryButton)} href={toTelHref(HEADER_PHONE_MAIN)}>
                <IconPhone currentColor="currentColor" height={14} width={14} />
                Связаться
              </a>
            </div>
          </section>
        </div>
      </div>
    </Page>
  );
};
