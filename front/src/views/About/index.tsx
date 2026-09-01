'use client';

import clsx from 'clsx';

import { useContent } from '@/core/entities/content';
import { IconLocation, IconPhone } from '@/core/shared/icons';
import { HEADER_PHONE_MAIN } from '@/core/shared/ui/Header/constants';
import { HtmlContent } from '@/core/shared/ui/HtmlContent';
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

const ORDER_EMAIL = 'order@voint.ru';

const toTelHref = (phone: string): string => `tel:${phone.replace(/[^\d+]/g, '')}`;

export const About = (): JSX.Element => {
  const { content, error, isPending } = useContent();
  const about = content?.about;
  const title = about?.title ?? 'О компании';
  const heroBadge = about?.hero_badge || ABOUT_HERO_BADGE;
  const heroSubtitle = about?.hero_subtitle || ABOUT_HERO_SUBTITLE;
  const profileBadge = about?.profile_badge || ABOUT_PROFILE_BADGE;
  const profileTitle = about?.profile_title || ABOUT_PROFILE_TITLE;
  const directionsBadge = about?.directions_badge || ABOUT_DIRECTIONS_BADGE;
  const directionsTitle = about?.directions_title || ABOUT_DIRECTIONS_TITLE;
  const directionsSubtitle = about?.directions_subtitle || ABOUT_DIRECTIONS_SUBTITLE;
  const officesBadge = about?.offices_badge || ABOUT_OFFICES_BADGE;
  const officesTitle = about?.offices_title || ABOUT_OFFICES_TITLE;
  const officesSubtitle = about?.offices_subtitle || ABOUT_OFFICES_SUBTITLE;
  const ctaBadge = about?.cta_badge || ABOUT_CTA_BADGE;
  const ctaTitle = about?.cta_title || ABOUT_CTA_TITLE;
  const ctaText = about?.cta_text || ABOUT_CTA_TEXT;
  const ctaButton = about?.cta_button || 'Отправить заявку';
  const ctaHint =
    about?.cta_hint ||
    'Вы можете отправить заявку по электронной почте на order@voint.ru, либо здесь:';

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <section className={clsx(styles.hero)}>
          <div className={clsx(styles.heroInner)}>
            <p className={clsx(styles.heroBadge)}>{heroBadge}</p>
            <h1 className={clsx(styles.heroTitle)}>{title}</h1>
            <p className={clsx(styles.heroDescription)}>{heroSubtitle}</p>
          </div>
        </section>

        <div className={clsx(styles.container)}>
          <section aria-labelledby="about-profile-title" className={clsx(styles.profile)}>
            <div className={clsx(styles.sectionIntro)}>
              <p className={clsx(styles.sectionBadge)}>{profileBadge}</p>
              <h2 className={clsx(styles.sectionTitle)} id="about-profile-title">
                {profileTitle}
              </h2>
            </div>

            {isPending && !about ? <p className={clsx(styles.status)}>Загрузка…</p> : null}
            {error && !about ? <p className={clsx(styles.error)}>{error}</p> : null}

            {about?.text ? (
              <HtmlContent className={clsx(styles.profileText)} text={about.text} />
            ) : null}
          </section>

          <section aria-labelledby="about-directions-title" className={clsx(styles.section)}>
            <div className={clsx(styles.sectionIntro)}>
              <p className={clsx(styles.sectionBadge)}>{directionsBadge}</p>
              <h2 className={clsx(styles.sectionTitle)} id="about-directions-title">
                {directionsTitle}
              </h2>
              <p className={clsx(styles.sectionSubtitle)}>{directionsSubtitle}</p>
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
              <p className={clsx(styles.sectionBadge)}>{officesBadge}</p>
              <h2 className={clsx(styles.sectionTitle)} id="about-offices-title">
                {officesTitle}
              </h2>
              <p className={clsx(styles.sectionSubtitle)}>{officesSubtitle}</p>
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
              <p className={clsx(styles.sectionBadge)}>{ctaBadge}</p>
              <h2 className={clsx(styles.ctaTitle)} id="about-cta-title">
                {ctaTitle}
              </h2>
              <p className={clsx(styles.ctaText)}>{ctaText}</p>
              <p className={clsx(styles.ctaHint)}>{ctaHint}</p>
            </div>
            <div className={clsx(styles.ctaActions)}>
              <a className={clsx(styles.primaryButton)} href={`mailto:${ORDER_EMAIL}`}>
                {ctaButton}
              </a>
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
