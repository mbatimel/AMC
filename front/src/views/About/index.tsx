'use client';

import clsx from 'clsx';

import { useContent } from '@/core/entities/content';
import { IconLocation, IconPhone } from '@/core/shared/icons';
import { HEADER_PHONE_MAIN } from '@/core/shared/ui/Header/constants';
import { HtmlContent } from '@/core/shared/ui/HtmlContent';
import { Page } from '@/core/shared/ui/Page';

import styles from './About.module.css';
import { ABOUT_DIRECTIONS } from './lib/aboutData';

const ORDER_EMAIL = 'order@voint.ru';

const toTelHref = (phone: string): string => `tel:${phone.replace(/[^\d+]/g, '')}`;

const AboutSkeleton = (): JSX.Element => (
  <div aria-busy="true" className={clsx(styles.root)} role="status">
    <section className={clsx(styles.hero)}>
      <div className={clsx(styles.heroInner)}>
        <span className={clsx(styles.skeletonLine, styles.skeletonBadge)} />
        <span className={clsx(styles.skeletonLine, styles.skeletonTitle)} />
        <span className={clsx(styles.skeletonLine, styles.skeletonSubtitle)} />
      </div>
    </section>
    <div className={clsx(styles.container)}>
      <div className={clsx(styles.skeletonBlock)}>
        <span className={clsx(styles.skeletonLine, styles.skeletonWide)} />
        <span className={clsx(styles.skeletonLine, styles.skeletonWide)} />
        <span className={clsx(styles.skeletonLine)} />
      </div>
    </div>
  </div>
);

export const About = (): JSX.Element => {
  const { content, error, isPending } = useContent();
  const about = content?.about;

  if (isPending && !about) {
    return (
      <Page>
        <AboutSkeleton />
      </Page>
    );
  }

  if (error && !about) {
    return (
      <Page>
        <div className={clsx(styles.root)}>
          <div className={clsx(styles.container)}>
            <p className={clsx(styles.error)}>{error}</p>
          </div>
        </div>
      </Page>
    );
  }

  if (!about) {
    return (
      <Page>
        <div className={clsx(styles.root)}>
          <div className={clsx(styles.container)}>
            <p className={clsx(styles.status)}>Контент страницы пока не опубликован.</p>
          </div>
        </div>
      </Page>
    );
  }

  const offices = (about.offices ?? []).map((office) => ({
    city: office.city,
    description: office.description,
    isMain: Boolean(office.is_main),
  }));

  return (
    <Page>
      <div className={clsx(styles.root)}>
        <section className={clsx(styles.hero)}>
          <div className={clsx(styles.heroInner)}>
            {about.hero_badge ? <p className={clsx(styles.heroBadge)}>{about.hero_badge}</p> : null}
            {about.title ? <h1 className={clsx(styles.heroTitle)}>{about.title}</h1> : null}
            {about.hero_subtitle ? (
              <p className={clsx(styles.heroDescription)}>{about.hero_subtitle}</p>
            ) : null}
          </div>
        </section>

        <div className={clsx(styles.container)}>
          <section aria-labelledby="about-profile-title" className={clsx(styles.profile)}>
            <div className={clsx(styles.sectionIntro)}>
              {about.profile_badge ? (
                <p className={clsx(styles.sectionBadge)}>{about.profile_badge}</p>
              ) : null}
              {about.profile_title ? (
                <h2 className={clsx(styles.sectionTitle)} id="about-profile-title">
                  {about.profile_title}
                </h2>
              ) : null}
            </div>

            {about.text ? (
              <HtmlContent className={clsx(styles.profileText)} text={about.text} />
            ) : null}
          </section>

          <section aria-labelledby="about-directions-title" className={clsx(styles.section)}>
            <div className={clsx(styles.sectionIntro)}>
              {about.directions_badge ? (
                <p className={clsx(styles.sectionBadge)}>{about.directions_badge}</p>
              ) : null}
              {about.directions_title ? (
                <h2 className={clsx(styles.sectionTitle)} id="about-directions-title">
                  {about.directions_title}
                </h2>
              ) : null}
              {about.directions_subtitle ? (
                <p className={clsx(styles.sectionSubtitle)}>{about.directions_subtitle}</p>
              ) : null}
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
              {about.offices_badge ? (
                <p className={clsx(styles.sectionBadge)}>{about.offices_badge}</p>
              ) : null}
              {about.offices_title ? (
                <h2 className={clsx(styles.sectionTitle)} id="about-offices-title">
                  {about.offices_title}
                </h2>
              ) : null}
              {about.offices_subtitle ? (
                <p className={clsx(styles.sectionSubtitle)}>{about.offices_subtitle}</p>
              ) : null}
            </div>

            {offices.length > 0 ? (
              <div className={clsx(styles.officesGrid)}>
                {offices.map((office) => (
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
            ) : null}
          </section>

          <section aria-labelledby="about-cta-title" className={clsx(styles.cta)}>
            <div className={clsx(styles.ctaCopy)}>
              {about.cta_badge ? (
                <p className={clsx(styles.sectionBadge)}>{about.cta_badge}</p>
              ) : null}
              {about.cta_title ? (
                <h2 className={clsx(styles.ctaTitle)} id="about-cta-title">
                  {about.cta_title}
                </h2>
              ) : null}
              {about.cta_text ? <p className={clsx(styles.ctaText)}>{about.cta_text}</p> : null}
              {about.cta_hint ? <p className={clsx(styles.ctaHint)}>{about.cta_hint}</p> : null}
            </div>
            <div className={clsx(styles.ctaActions)}>
              {about.cta_button ? (
                <a className={clsx(styles.primaryButton)} href={`mailto:${ORDER_EMAIL}`}>
                  {about.cta_button}
                </a>
              ) : null}
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
