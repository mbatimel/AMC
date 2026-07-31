import clsx from 'clsx';
import Link from 'next/link';

import { AppPath } from '@/core/shared/router/paths';

import styles from './InfoPage.module.css';

type InfoPageProps = {
  children: React.ReactNode;
  description?: string;
  eyebrow?: string;
  title: string;
};

export const InfoPage = ({
  children,
  description,
  eyebrow,
  title,
}: InfoPageProps): JSX.Element => {
  return (
    <div className={clsx(styles.root)}>
      <section className={clsx(styles.hero)}>
        <div className={clsx(styles.heroInner)}>
          {eyebrow ? <p className={clsx(styles.eyebrow)}>{eyebrow}</p> : null}
          <h1 className={clsx(styles.title)}>{title}</h1>
          {description ? <p className={clsx(styles.description)}>{description}</p> : null}
        </div>
      </section>

      <div className={clsx(styles.container)}>
        <nav aria-label="Хлебные крошки" className={clsx(styles.breadcrumbs)}>
          <Link href={AppPath.Home}>Главная</Link>
          <span>/</span>
          <span>{title}</span>
        </nav>
        {children}
      </div>
    </div>
  );
};

type InfoCardProps = {
  action?: React.ReactNode;
  children: React.ReactNode;
  title?: string;
};

export const InfoCard = ({ action, children, title }: InfoCardProps): JSX.Element => {
  return (
    <section className={clsx(styles.card)}>
      {title || action ? (
        <header className={clsx(styles.cardHeader)}>
          {title ? <h2 className={clsx(styles.cardTitle)}>{title}</h2> : <span />}
          {action}
        </header>
      ) : null}
      <div className={clsx(styles.cardBody)}>{children}</div>
    </section>
  );
};

type InfoTextProps = {
  text: string;
};

/** Многострочный текст из редактора контента — абзацы по пустой строке. */
export const InfoText = ({ text }: InfoTextProps): JSX.Element => {
  const paragraphs = text.split('\n').filter((line) => line.trim().length > 0);

  return (
    <>
      {paragraphs.map((paragraph, index) => (
        <p className={clsx(styles.paragraph)} key={`${index}-${paragraph.slice(0, 12)}`}>
          {paragraph}
        </p>
      ))}
    </>
  );
};

type InfoListProps = {
  items: string[];
};

export const InfoList = ({ items }: InfoListProps): JSX.Element => {
  return (
    <ul className={clsx(styles.list)}>
      {items.map((item, index) => (
        <li className={clsx(styles.listItem)} key={`${index}-${item.slice(0, 12)}`}>
          {item}
        </li>
      ))}
    </ul>
  );
};

export const InfoPageSkeleton = (): JSX.Element => (
  <div className={clsx(styles.skeleton)} role="status">
    <span className={clsx(styles.skeletonLine)} />
    <span className={clsx(styles.skeletonLine)} />
    <span className={clsx(styles.skeletonLine)} />
  </div>
);
