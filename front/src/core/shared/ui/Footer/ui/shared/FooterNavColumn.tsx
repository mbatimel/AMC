import clsx from 'clsx';
import Link from 'next/link';

import type { FooterLinkItem } from '../../constants';

import styles from '../../Footer.module.css';

type FooterNavColumnProps = {
  items: FooterLinkItem[];
  title: string;
};

export const FooterNavColumn = ({ items, title }: FooterNavColumnProps): JSX.Element => {
  return (
    <nav aria-label={title}>
      <div className={clsx(styles.navTitle)}>
        <span>{title}</span>
      </div>

      <ul className={clsx(styles.navList)}>
        {items.map((item) => (
          <li className={clsx(styles.navItem)} key={item.label}>
            <Link className={clsx(styles.navLink)} href={item.href}>
              <span>{item.label}</span>
            </Link>
          </li>
        ))}
      </ul>
    </nav>
  );
};
